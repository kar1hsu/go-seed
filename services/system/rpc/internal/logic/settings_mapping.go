package logic

import (
	"github.com/kar1hsu/go-seed/pkg/xerr"
	"github.com/kar1hsu/go-seed/services/system/internal/application/management"
	"github.com/kar1hsu/go-seed/services/system/internal/model"
	"github.com/kar1hsu/go-seed/services/system/rpc/pb"
	"google.golang.org/grpc/codes"
	"strconv"
)

func dictionaryRecord(d model.Dictionary) *pb.DictionaryRecord {
	return &pb.DictionaryRecord{Id: strconv.FormatUint(d.ID, 10), Code: d.Code, Name: d.Name, Description: d.Description, Enabled: d.Enabled, Revision: management.Revision(d.UpdatedAt), CreatedAt: management.Revision(d.CreatedAt), ItemCount: d.ItemCount}
}
func dictionaryItems(items []*pb.DictionaryItem) ([]model.DictionaryItem, error) {
	result := []model.DictionaryItem{}
	for _, item := range items {
		if item == nil {
			return nil, xerr.New(codes.InvalidArgument)
		}
		result = append(result, model.DictionaryItem{Value: item.Value, Label: item.Label, Sort: item.Sort, Enabled: item.Enabled})
	}
	return result, nil
}
func dictionaryResponse(d *management.DictionaryDetail) *pb.DictionaryResponse {
	result := &pb.DictionaryResponse{Record: dictionaryRecord(d.Record), Items: []*pb.DictionaryItem{}}
	for _, item := range d.Items {
		result.Items = append(result.Items, &pb.DictionaryItem{Value: item.Value, Label: item.Label, Sort: item.Sort, Enabled: item.Enabled})
	}
	return result
}
func settingRecord(s model.Setting) *pb.SettingRecord {
	return &pb.SettingRecord{Id: strconv.FormatUint(s.ID, 10), Code: s.Code, Name: s.Name, Description: s.Description, Type: s.Type, Value: s.Value, Enabled: s.Enabled, Revision: management.Revision(s.UpdatedAt), CreatedAt: management.Revision(s.CreatedAt)}
}
