package live_http

import (
	"gil_teacher/app/dao/live_room"
	"gil_teacher/libs/timex"

	"github.com/jinzhu/copier"
)

func liveRoom2LiveRoomVo(inList []live_room.LiveRoom) (outList []LiveRoomVo) {

	if len(inList) <= 0 {
		return outList
	}

	var result []LiveRoomVo
	for _, it := range inList {
		var vo LiveRoomVo
		_ = copier.Copy(&vo, &it)

		// TODO 其它

		vo.UpdatedAt = timex.TimeFormat(it.UpdatedAt, timex.DefaultLayout)
		vo.CreatedAt = timex.TimeFormat(it.CreatedAt, timex.DefaultLayout)

		result = append(result, vo)
	}

	return result

}
