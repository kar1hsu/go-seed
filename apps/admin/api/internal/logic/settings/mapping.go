package settings

import (
	"github.com/kar1hsu/go-seed/apps/admin/api/internal/types"
	"github.com/kar1hsu/go-seed/services/system/rpc/pb"
)

func dictionaryRecord(d *pb.DictionaryRecord) types.DictionaryRecord {
	return types.DictionaryRecord{Id: d.Id, Code: d.Code, Name: d.Name, Description: d.Description, Enabled: d.Enabled, Revision: d.Revision, CreatedAt: d.CreatedAt, ItemCount: d.ItemCount}
}
func dictionaryItems(items []types.DictionaryItem) []*pb.DictionaryItem {
	result := []*pb.DictionaryItem{}
	for _, item := range items {
		result = append(result, &pb.DictionaryItem{Value: item.Value, Label: item.Label, Sort: item.Sort, Enabled: item.Enabled})
	}
	return result
}
func dictionaryDetail(d *pb.DictionaryResponse) types.DictionaryDetailData {
	result := types.DictionaryDetailData{Record: dictionaryRecord(d.Record), Items: []types.DictionaryItem{}}
	for _, item := range d.Items {
		result.Items = append(result.Items, types.DictionaryItem{Value: item.Value, Label: item.Label, Sort: item.Sort, Enabled: item.Enabled})
	}
	return result
}
func settingRecord(s *pb.SettingRecord) types.SettingRecord {
	return types.SettingRecord{Id: s.Id, Code: s.Code, Name: s.Name, Description: s.Description, Type: s.Type, Value: s.Value, Enabled: s.Enabled, Revision: s.Revision, CreatedAt: s.CreatedAt}
}
