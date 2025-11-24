package dbs

import (
	"context"
	"fmt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"sync"
)

func InsertOrUpdate(ctx context.Context, db *gorm.DB, data interface{}) *gorm.DB {
	return db.WithContext(ctx).Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Create(data)
}

func InsertOrNothing(ctx context.Context, db *gorm.DB, data interface{}) *gorm.DB {
	return db.WithContext(ctx).Clauses(clause.OnConflict{
		DoNothing: true,
	}).Create(data)
}

type RecordsResult[T any] struct {
	Data  []T
	Size  int
	Total int
}

type singleResult[T any] struct {
	dbInfo DBInfo
	data   []T
	err    error
}

type countResult struct {
	dbInfo DBInfo
	count  int
	err    error
}

var notFoundCluster = "未找到数据源:[%v]"

func SelectAllDataByCluster[T any](ctx context.Context, cluster DBClusterEnum, t T, sql string, params interface{}, autoInjectDBInfo bool) (recordsResult *RecordsResult[T], err error) {
	clusterDBS := GetDBByCluster(cluster)
	if len(clusterDBS) == 0 {
		return nil, fmt.Errorf(notFoundCluster, cluster)
	}
	var singleResultCh = make(chan singleResult[T], len(clusterDBS))
	var wg sync.WaitGroup
	for _, clusterdb := range clusterDBS {
		wg.Add(1)
		go selectSingleDBFunc(ctx, singleResultCh, &wg, sql, clusterdb, params, autoInjectDBInfo)
	}
	wg.Wait()
	close(singleResultCh)
	chunks := make([][]T, 0, len(clusterDBS))
	var resultTotal = 0
	for ch := range singleResultCh {
		if ch.err != nil {
			return nil, ch.err
		}
		resultTotal = resultTotal + len(ch.data)
		chunks = append(chunks, ch.data)
	}
	if resultTotal == 0 {
		return &RecordsResult[T]{
			Data:  nil,
			Size:  0,
			Total: 0,
		}, nil
	}
	var mergeData = make([]T, 0, resultTotal)
	for _, chk := range chunks {
		mergeData = append(mergeData, chk...)
	}
	return &RecordsResult[T]{
		Data:  mergeData,
		Size:  resultTotal,
		Total: resultTotal,
	}, nil
}
func selectSingleDBFunc[T any](ctx context.Context, singleResultCh chan singleResult[T], wg *sync.WaitGroup, sql string, dbInfo DBInfo, params interface{}, autoInjectDBInfo bool) {
	defer wg.Done()
	var queryData []T
	tx := autoExecSql(ctx, dbInfo, sql, params, &queryData, autoInjectDBInfo)
	if tx.Error != nil {
		singleResultCh <- singleResult[T]{
			dbInfo: dbInfo,
			data:   nil,
			err:    tx.Error,
		}
		return
	} else {
		singleResultCh <- singleResult[T]{
			dbInfo: dbInfo,
			data:   queryData,
			err:    nil,
		}
	}

}

func SelectPageDataByCluster[T any](ctx context.Context, cluster DBClusterEnum, t T, sqlCount string, sqlSelect string, params interface{}, limit int, offset int, autoInjectDBInfo bool) (*RecordsResult[T], error) {
	var wg sync.WaitGroup
	clusterDBS := GetDBByCluster(cluster)
	if len(clusterDBS) == 0 {
		return nil, fmt.Errorf(notFoundCluster, cluster)
	}
	var countCh = make(chan countResult, len(clusterDBS))
	for _, info := range clusterDBS {
		wg.Add(1)
		go selectCountFunc(ctx, countCh, &wg, info, sqlCount, params, false)
	}
	wg.Wait()
	close(countCh)

	var selectResult = make([]T, 0, limit)
	rawSql := fmt.Sprintf("%s limit @limit offset @offset", sqlSelect)
	pageMap := map[string]interface{}{
		"limit":  limit,
		"offset": offset,
	}
	var total int
	for ch := range countCh {
		if ch.err != nil {
			return nil, ch.err
		}

		if ch.count == 0 {
			continue
		}
		total = total + ch.count
		if pageMap["limit"].(int) <= 0 {
			continue
		}
		if pageMap["offset"].(int) > ch.count {
			pageMap["offset"] = pageMap["offset"].(int) - ch.count
			continue
		}
		var selectResultTemp []T

		tx := autoExecPageSql(ctx, ch.dbInfo, rawSql, params, pageMap, &selectResultTemp, autoInjectDBInfo)
		if tx.Error != nil {
			return nil, tx.Error
		}
		selectResult = append(selectResult, selectResultTemp...)
		num := len(selectResult)
		pageMap["limit"] = pageMap["limit"].(int) - num
		pageMap["offset"] = 0
	}
	return &RecordsResult[T]{
		Data:  selectResult,
		Size:  len(selectResult),
		Total: total,
	}, nil
}

func autoExecPageSql(ctx context.Context, dbInfo DBInfo, sql string, params interface{}, pageMap map[string]interface{}, result interface{}, autoInjectDBInfo bool) *gorm.DB {
	var tx *gorm.DB
	if params != nil && autoInjectDBInfo {
		tx = dbInfo.DB.WithContext(ctx).Raw(sql, params, pageMap, dbInfo).Scan(result)
	} else if params != nil && !autoInjectDBInfo {
		tx = dbInfo.DB.WithContext(ctx).Raw(sql, params, pageMap).Scan(result)
	} else if params == nil && autoInjectDBInfo {
		tx = dbInfo.DB.WithContext(ctx).Raw(sql, dbInfo, pageMap).Scan(result)
	} else {
		tx = dbInfo.DB.WithContext(ctx).Raw(sql, pageMap).Scan(result)
	}
	return tx
}

func autoExecSql(ctx context.Context, dbInfo DBInfo, sql string, params interface{}, result interface{}, autoInjectDBInfo bool) *gorm.DB {
	var tx *gorm.DB
	if params != nil && autoInjectDBInfo {
		tx = dbInfo.DB.WithContext(ctx).Raw(sql, params, dbInfo).Scan(result)
	} else if params != nil && !autoInjectDBInfo {
		tx = dbInfo.DB.WithContext(ctx).Raw(sql, params).Scan(result)
	} else if params == nil && autoInjectDBInfo {
		tx = dbInfo.DB.WithContext(ctx).Raw(sql, dbInfo).Scan(result)
	} else {
		tx = dbInfo.DB.WithContext(ctx).Raw(sql).Scan(result)
	}
	return tx
}
func selectCountFunc(ctx context.Context, countCh chan countResult, wg *sync.WaitGroup, dbInfo DBInfo, sqlCount string, params interface{}, autoInjectDBInfo bool) {
	defer wg.Done()
	var num = 0
	tx := autoExecSql(ctx, dbInfo, sqlCount, params, &num, autoInjectDBInfo)
	if tx.Error != nil {
		countCh <- countResult{
			dbInfo: dbInfo,
			count:  0,
			err:    tx.Error,
		}
	} else {
		countCh <- countResult{
			dbInfo: dbInfo,
			count:  num,
			err:    nil,
		}
	}
}
