package sqldb

import (
	"context"
	"errors"
	"fmt"
	"platform-go-challenge/domain"
)

func (d *DB) AddAsset(ctx context.Context, asset domain.InputAsset) (*domain.Asset, error) {
	newAsset := &domain.Asset{}
	switch v := asset.Data.(type) {
	case *domain.Insight:
		in := &Insight{}
		in.FromDomain(v)
		err := d.db.Create(in).Error
		if err != nil {
			return nil, err
		}
		newAsset.ID = in.ID
		newAsset.Data = in.ToDomain()
	case *domain.Chart:
		ch := &Chart{}
		ch.FromDomain(v)
		err := d.db.Create(ch).Error
		if err != nil {
			return nil, err
		}
		newAsset.ID = ch.ID
		newAsset.Data = ch.ToDomain()
	case *domain.Audience:
		au := &Audience{}
		au.FromDomain(v)
		err := d.db.Create(au).Error
		if err != nil {
			return nil, err
		}
		newAsset.ID = au.ID
		newAsset.Data = au.ToDomain()
	default:
		return nil, fmt.Errorf("AddAsset: %w", ErrThisAssetTypeDoesNotExist)
	}
	return newAsset, nil
}

func (d *DB) UpdateAsset(ctx context.Context, assetID uint, asset domain.InputAsset) (*domain.Asset, error) {
	if assetID <= 0 {
		return nil, errors.New("add id ")
	}
	newAsset := &domain.Asset{}
	switch v := asset.Data.(type) {
	case *domain.Insight:
		in := &Insight{}
		d.db.First(in, assetID)
		in.FromDomain(v)
		err := d.db.Save(in).Error
		if err != nil {
			return nil, err
		}
		newAsset.ID = in.ID
		newAsset.Data = in.ToDomain()
	case *domain.Chart:
		ch := &Chart{}
		d.db.First(ch, assetID)
		ch.FromDomain(v)
		err := d.db.Save(ch).Error
		if err != nil {
			return nil, err
		}
		newAsset.ID = ch.ID
		newAsset.Data = ch.ToDomain()

	case *domain.Audience:
		au := &Audience{}
		d.db.First(au, assetID)
		au.FromDomain(v)
		err := d.db.Save(au).Error
		if err != nil {
			return nil, err
		}
		newAsset.ID = au.ID
		newAsset.Data = au.ToDomain()
	default:
		return nil, fmt.Errorf("UpdateAsset: %w", ErrThisAssetTypeDoesNotExist)
	}

	return newAsset, nil
}

func (d *DB) GetAsset(ctx context.Context, at domain.AssetType, assetID uint) (*domain.Asset, error) {
	newAsset := &domain.Asset{}
	switch at {
	case domain.InsightAssetType:
		in := &Insight{}
		err := d.db.First(in, assetID).Error
		if err != nil {
			return nil, err
		}
		newAsset.ID = in.ID
		newAsset.Data = in.ToDomain()
	case domain.ChartAssetType:
		ch := &Chart{}
		err := d.db.First(ch, assetID).Error
		if err != nil {
			return nil, err
		}
		newAsset.ID = ch.ID
		newAsset.Data = ch.ToDomain()
	case domain.AudienceAssetType:
		au := &Audience{}
		err := d.db.First(au, assetID).Error
		if err != nil {
			return nil, err
		}
		newAsset.ID = au.ID
		newAsset.Data = au.ToDomain()
	default:
		return nil, fmt.Errorf("GetAsset: %w", ErrThisAssetTypeDoesNotExist)
	}
	return newAsset, nil
}

func (d *DB) DeleteAsset(ctx context.Context, at domain.AssetType, assetID uint) error {
	switch at {
	case domain.InsightAssetType:
		err := d.db.Unscoped().Delete(&Insight{}, assetID).Error
		if err != nil {
			return err
		}

	case domain.ChartAssetType:
		err := d.db.Unscoped().Delete(&Chart{}, assetID).Error
		if err != nil {
			return err
		}

	case domain.AudienceAssetType:
		err := d.db.Unscoped().Delete(&Audience{}, assetID).Error
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("DeleteAsset: %w", ErrThisAssetTypeDoesNotExist)
	}
	return nil
}

func (d *DB) ListAssets(ctx context.Context, query domain.QueryAssets) (*domain.ListedAssets, error) {
	gormQuery := d.db
	if query.IsDesc {
		gormQuery = gormQuery.Where("id < ?", query.LastID).Order("id desc")
	} else {
		gormQuery = gormQuery.Where("id > ?", query.LastID)
	}
	assets := []domain.Asset{}
	switch query.Type {
	case domain.InsightAssetType:
		ins := []Insight{}
		err := gormQuery.Limit(query.Limit).Find(&ins).Error
		if err != nil {
			return nil, err
		}
		assets = listRowsToAssets(ins)
	case domain.ChartAssetType:
		chs := []Chart{}
		err := gormQuery.Limit(query.Limit).Find(&chs).Error
		if err != nil {
			return nil, err
		}
		assets = listRowsToAssets(chs)
	case domain.AudienceAssetType:
		aus := []Audience{}
		err := gormQuery.Limit(query.Limit).Find(&aus).Error
		if err != nil {
			return nil, err
		}
		assets = listRowsToAssets(aus)
	}
	var firstID uint = 0
	var lastID uint = 0
	if len(assets) > 0 {
		firstID = uint(assets[0].ID)
		lastID = uint(assets[len(assets)-1].ID)
	}

	dl := domain.ListedAssets{
		FirstID: firstID,
		LastID:  lastID,
		Limit:   query.Limit,
		Assets:  assets,
	}
	return &dl, nil
}

func (d *DB) FavouriteAsset(ctx context.Context, userID, assetID uint, at domain.AssetType, isFavourite bool) (uint, error) {
	var nid uint = 0
	var err error
	switch at {
	case domain.InsightAssetType:
		if isFavourite {
			count := int64(0)
			d.db.Model(FavouriteInsight{}).Where("user_id = ? AND insight_id = ? ", userID, assetID).Count(&count)
			if count == 0 {
				in := &FavouriteInsight{UserID: userID, InsightID: assetID}
				err = d.db.Create(in).Error
				nid = in.ID
			} else {
				err = errors.New("record exists")
			}
		} else {
			in := &FavouriteInsight{}
			err = d.db.Where("user_id = ? AND insight_id = ? ", userID, assetID).Unscoped().Delete(in).Error
		}
	case domain.ChartAssetType:
		if isFavourite {
			count := int64(0)
			d.db.Model(FavouriteChart{}).Where("user_id = ? AND chart_id = ? ", userID, assetID).Count(&count)
			if count == 0 {
				ch := &FavouriteChart{UserID: userID, ChartID: assetID}
				err = d.db.Create(ch).Error
				nid = ch.ID
			} else {
				err = errors.New("record exists")
			}
		} else {
			in := &FavouriteChart{}
			err = d.db.Where("user_id = ? AND chart_id = ? ", userID, assetID).Unscoped().Delete(in).Error
		}
	case domain.AudienceAssetType:
		if isFavourite {
			count := int64(0)
			d.db.Model(FavouriteAudience{}).Where("user_id = ? AND audience_id = ? ", userID, assetID).Count(&count)
			if count == 0 {
				au := &FavouriteAudience{UserID: userID, AudienceID: assetID}
				err = d.db.Create(au).Error
				nid = au.ID
			} else {
				err = errors.New("record exists")
			}
		} else {
			in := &FavouriteAudience{}
			err = d.db.Where("user_id = ? AND audience_id = ? ", userID, assetID).Unscoped().Delete(in).Error
		}
	}
	if err != nil {
		return 0, err
	}

	return nid, nil
}

func (d *DB) listFavouriteAudiences(ctx context.Context, userID uint, onlyFav bool, query domain.QueryAssets) (*domain.ListedAssets, error) {
	var aus []AudienceWithFavour
	gormQuery := d.db.Model(Audience{}).Select("audiences.*, (favourite_audiences.user_id = ?) AS is_favourite", userID)
	if onlyFav {
		if query.IsDesc {
			gormQuery = gormQuery.Joins("INNER JOIN favourite_audiences ON favourite_audiences.audience_id = audiences.id AND audiences.id < ? AND favourite_audiences.user_id = ?", query.LastID, userID).Order("audiences.id desc")
		} else {
			gormQuery = gormQuery.Joins("INNER JOIN favourite_audiences ON favourite_audiences.audience_id = audiences.id AND audiences.id > ? AND favourite_audiences.user_id = ?", query.LastID, userID).Order("audiences.id asc")
		}
	} else {
		if query.IsDesc {
			gormQuery = gormQuery.Joins("LEFT JOIN favourite_audiences ON favourite_audiences.audience_id = audiences.id AND audiences.id < ? AND favourite_audiences.user_id = ?", query.LastID, userID).Order("audiences.id desc")
		} else {
			gormQuery = gormQuery.Joins("LEFT JOIN favourite_audiences ON favourite_audiences.audience_id = audiences.id AND audiences.id > ? AND favourite_audiences.user_id = ?", query.LastID, userID).Order("audiences.id asc")
		}
	}

	gormQuery.Limit(query.Limit).Unscoped().Find(&aus)
	assets := listRowsToAssets(aus)
	var firstID uint = 0
	var lastID uint = 0
	if len(assets) > 0 {
		firstID = uint(assets[0].ID)
		lastID = uint(assets[len(assets)-1].ID)
	}
	la := domain.ListedAssets{
		FirstID: firstID,
		LastID:  lastID,
		Limit:   query.Limit,
		Type:    query.Type,
		Assets:  assets,
	}

	return &la, nil
}

func (d *DB) listFavouriteInsights(ctx context.Context, userID uint, onlyFav bool, query domain.QueryAssets) (*domain.ListedAssets, error) {
	var ins []InsightWithFavour
	gormQuery := d.db.Model(Insight{}).Select("insights.*, (favourite_insights.user_id = ?) AS is_favourite", userID)
	if onlyFav {
		if query.IsDesc {
			gormQuery = gormQuery.Joins("INNER JOIN favourite_insights ON favourite_insights.insight_id = insights.id AND insights.id < ? AND favourite_insights.user_id = ?", query.LastID, userID).Order("insights.id desc")
		} else {
			gormQuery = gormQuery.Joins("INNER JOIN favourite_insights ON favourite_insights.insight_id = insights.id AND insights.id > ? AND favourite_insights.user_id = ?", query.LastID, userID).Order("insights.id asc")
		}
	} else {
		if query.IsDesc {
			gormQuery = gormQuery.Joins("LEFT JOIN favourite_insights ON favourite_insights.insight_id = insights.id AND insights.id < ? AND favourite_insights.user_id = ?", query.LastID, userID).Order("insights.id desc")
		} else {
			gormQuery = gormQuery.Joins("LEFT JOIN favourite_insights ON favourite_insights.insight_id = insights.id AND insights.id > ? AND favourite_insights.user_id = ?", query.LastID, userID).Order("insights.id asc")
		}
	}

	gormQuery.Limit(query.Limit).Unscoped().Find(&ins)
	assets := listRowsToAssets(ins)
	var firstID uint = 0
	var lastID uint = 0
	if len(assets) > 0 {
		firstID = uint(assets[0].ID)
		lastID = uint(assets[len(assets)-1].ID)
	}
	la := domain.ListedAssets{
		FirstID: firstID,
		LastID:  lastID,
		Limit:   query.Limit,
		Type:    query.Type,
		Assets:  assets,
	}

	return &la, nil
}

func (d *DB) listFavouriteCharts(ctx context.Context, userID uint, onlyFav bool, query domain.QueryAssets) (*domain.ListedAssets, error) {
	var chs []ChartWithFavour
	gormQuery := d.db.Model(Chart{}).Select("charts.*, (favourite_charts.user_id = ?) AS is_favourite", userID)
	if onlyFav {
		if query.IsDesc {
			gormQuery = gormQuery.Joins("INNER JOIN favourite_charts ON favourite_charts.chart_id = charts.id AND charts.id < ? AND favourite_charts.user_id = ?", query.LastID, userID).Order("charts.id desc")
		} else {
			gormQuery = gormQuery.Joins("INNER JOIN favourite_charts ON favourite_charts.chart_id = charts.id AND charts.id > ? AND favourite_charts.user_id = ?", query.LastID, userID).Order("charts.id asc")
		}
	} else {
		if query.IsDesc {
			gormQuery = gormQuery.Joins("LEFT JOIN favourite_charts ON favourite_charts.chart_id = charts.id AND charts.id < ? AND favourite_charts.user_id = ?", query.LastID, userID).Order("charts.id desc")
		} else {
			gormQuery = gormQuery.Joins("LEFT JOIN favourite_charts ON favourite_charts.chart_id = charts.id AND charts.id > ? AND favourite_charts.user_id = ?", query.LastID, userID).Order("charts.id asc")
		}
	}

	gormQuery.Limit(query.Limit).Unscoped().Find(&chs)
	assets := listRowsToAssets(chs)
	var firstID uint = 0
	var lastID uint = 0
	if len(assets) > 0 {
		firstID = uint(assets[0].ID)
		lastID = uint(assets[len(assets)-1].ID)
	}
	la := domain.ListedAssets{
		FirstID: firstID,
		LastID:  lastID,
		Limit:   query.Limit,
		Type:    query.Type,
		Assets:  assets,
	}

	return &la, nil
}

func (d *DB) ListFavouriteAssets(ctx context.Context, userID uint, onlyFav bool, query domain.QueryAssets) (*domain.ListedAssets, error) {
	switch query.Type {
	case domain.InsightAssetType:
		return d.listFavouriteInsights(ctx, userID, onlyFav, query)
	case domain.AudienceAssetType:
		return d.listFavouriteAudiences(ctx, userID, onlyFav, query)
	case domain.ChartAssetType:
		return d.listFavouriteCharts(ctx, userID, onlyFav, query)
	}
	return nil, fmt.Errorf("ListFavouriteAssets: %w", ErrThisAssetTypeDoesNotExist)
}

func (d *DB) RemoveFavouriteAssetFromEveryone(ctx context.Context, assetID uint, at domain.AssetType) error {
	switch at {
	case domain.InsightAssetType:
		err := d.db.Unscoped().Where("insight_id = ? ", assetID).Delete(&FavouriteInsight{}).Error
		if err != nil {
			return err
		}
	case domain.ChartAssetType:
		err := d.db.Unscoped().Where("chart_id = ? ", assetID).Delete(&FavouriteChart{}).Error
		if err != nil {
			return err
		}
	case domain.AudienceAssetType:
		err := d.db.Unscoped().Where("audience_id = ? ", assetID).Delete(&FavouriteAudience{}).Error
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("RemoveFavouriteAssetFromEveryone: %w", ErrThisAssetTypeDoesNotExist)
	}
	return nil
}
