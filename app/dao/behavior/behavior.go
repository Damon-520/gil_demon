package behavior

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"gil_teacher/app/consts"
	clogger "gil_teacher/app/core/logger"
	"gil_teacher/app/dao"
	"gil_teacher/app/model/dto"
	"gil_teacher/app/utils"
	"gil_teacher/app/utils/idtools"

	"github.com/pkg/errors"
)

// BehaviorDAO 行为数据访问接口
type BehaviorDAO interface {
	// 保存教师行为
	SaveTeacherBehavior(ctx context.Context, behaviors []*dto.TeacherBehaviorDTO) error
	// 保存学生行为
	SaveStudentBehavior(ctx context.Context, behaviors []*dto.StudentBehaviorDTO) error
	// 创建一个会话，返回会话 id
	OpenCommunicationSession(ctx context.Context, session *dto.CommunicationSessionDTO) (string, error)
	// 查询会话
	GetCommunicationSession(ctx context.Context, sessionID string) (*dto.CommunicationSessionDTO, error)
	// 保存会话记录
	SaveCommunication(ctx context.Context, sessions []*dto.CommunicationSessionDTO, messages []*dto.CommunicationMessageDTO) error
	// 关闭会话，更新表
	CloseCommunicationSession(ctx context.Context, sessionID string) error
	// 查询指定会话的全部消息
	GetCommunicationSessionMessages(ctx context.Context, sessionID string, pageInfo *consts.DBPageInfo) ([]*dto.CommunicationMessageDTO, error)
	// 查询某个会话指定 id 的消息列表
	GetCommunicationSessionMessagesByIDs(ctx context.Context, sessionID string, messageIDs []string) ([]*dto.CommunicationMessageDTO, error)
	// 查询指定课堂的全部消息
	GetClassroomMessages(ctx context.Context, classroomID string, pageInfo *consts.DBPageInfo) ([]*dto.CommunicationMessageDTO, error)
	// 获取班级学生最新行为
	GetClassLatestBehaviors(ctx context.Context, ClassroomID uint64) ([]*dto.StudentLatestBehaviorDTO, error)
	// 获取班级学生所有行为（用于汇总统计）
	GetClassAllBehaviors(ctx context.Context, ClassroomID uint64) ([]*dto.StudentLatestBehaviorDTO, error)
	// 获取学生课堂详情
	GetStudentClassroomDetail(ctx context.Context, studentID, classroomID uint64) (*dto.StudentClassroomDetailDTO, error)
	// 获取课堂行为分类
	GetClassBehaviorCategory(ctx context.Context, classID, classroomID uint64) (*dto.ClassBehaviorCategoryDTO, error)
	// GetStudentsBehaviors 获取指定学生ID列表的行为数据
	GetStudentsBehaviors(ctx context.Context, studentIDs []uint64) ([]*dto.StudentLatestBehaviorDTO, error)
	// CountStudentTaskPraiseAndAttention 统计学生任务点赞和关注
	CountStudentTaskPraiseAndAttention(ctx context.Context, taskID, assignID uint64, studentIDs []uint64) ([]dao.CHGroupCountResult, error)
	// 获取学生特定类型的行为数据
	GetStudentBehaviorsByType(ctx context.Context, studentID, classroomID uint64, behaviorType string) ([]*StudentBehavior, error)
}

// BehaviorDAOImpl 行为数据访问对象实现
type BehaviorDAOImpl struct {
	communicationSessionDao *CommunicationSessionDao
	communicationMessageDao *CommunicationMessageDao
	studentBehaviorDao      *StudentBehaviorDao
	teacherBehaviorDao      *TeacherBehaviorDao
	logger                  *clogger.ContextLogger
}

// NewBehaviorDAO 创建行为数据访问对象
func NewBehaviorDAO(chClients map[string]*dao.ClickHouseRWClient, logger *clogger.ContextLogger) BehaviorDAO {
	return &BehaviorDAOImpl{
		communicationSessionDao: newCommunicationSessionDao(chClients[consts.ChDBTeacher], logger),
		communicationMessageDao: newCommunicationMessageDao(chClients[consts.ChDBTeacher], logger),
		teacherBehaviorDao:      newTeacherBehaviorDao(chClients[consts.ChDBTeacher], logger),
		studentBehaviorDao:      newStudentBehaviorDao(chClients[consts.ChDBStudent], logger),
		logger:                  logger,
	}
}

// 创建一个会话，返回会话 id
func (d *BehaviorDAOImpl) OpenCommunicationSession(ctx context.Context, session *dto.CommunicationSessionDTO) (string, error) {
	sessionID := idtools.GetUUID()
	// session 需要转成 db 模型
	sessionModel := &CommunicationSession{
		SessionID:   sessionID,
		UserID:      session.UserID,
		UserType:    session.UserType,
		SchoolID:    session.SchoolID,
		CourseID:    utils.Ptr(session.CourseID),
		ClassroomID: utils.Ptr(session.ClassroomID),
		SessionType: session.SessionType,
		TargetID:    session.TargetID,
		StartTime:   time.Now(),
	}

	err := d.communicationSessionDao.SaveCommunicationSessions(ctx, []*CommunicationSession{sessionModel})
	if err != nil {
		d.logger.Error(ctx, "insert communication session failed, error: %v", err)
		return "", errors.Wrap(err, "insert communication session failed")
	}

	// 还需要记录到行为表中，根据用户类型记录
	if session.UserType == string(consts.CommunicationUserTypeStudent) {
		// 学生发起会话
		err = d.SaveStudentBehavior(ctx, []*dto.StudentBehaviorDTO{
			{
				SchoolID:               session.SchoolID,
				ClassID:                session.ClassID,
				ClassroomID:            &session.ClassroomID,
				StudentID:              session.UserID,
				BehaviorType:           consts.BehaviorTypeCommunication,
				CommunicationSessionID: &sessionID,
				CreateTime:             time.Now(),
			},
		})
		if err != nil {
			d.logger.Error(ctx, "save student behavior failed, error: %v", err)
			return "", errors.Wrap(err, "save student behavior failed")
		}
	} else if session.UserType == string(consts.CommunicationUserTypeTeacher) {
		// 老师发起会话
		err = d.SaveTeacherBehavior(ctx, []*dto.TeacherBehaviorDTO{
			{
				SchoolID:               session.SchoolID,
				ClassID:                session.ClassID,
				ClassroomID:            &session.ClassroomID,
				TeacherID:              session.UserID,
				BehaviorType:           consts.BehaviorTypeCommunication,
				CommunicationSessionID: &sessionID,
				CreateTime:             time.Now(),
			},
		})
		if err != nil {
			d.logger.Error(ctx, "save teacher behavior failed, error: %v", err)
			return "", errors.Wrap(err, "save teacher behavior failed")
		}
	}
	return sessionID, nil
}

// 查询会话
func (d *BehaviorDAOImpl) GetCommunicationSession(ctx context.Context, sessionID string) (*dto.CommunicationSessionDTO, error) {
	records, err := d.communicationSessionDao.GetSessionsByIDs(ctx, []string{sessionID})
	if err != nil {
		return nil, errors.Wrap(err, "find communication session failed")
	}
	if len(records) == 0 {
		return nil, errors.New("communication session not found")
	}

	session := records[0]
	return &dto.CommunicationSessionDTO{
		SessionID:   session.SessionID,
		UserID:      session.UserID,
		UserType:    session.UserType,
		SchoolID:    session.SchoolID,
		CourseID:    utils.PtrValue(session.CourseID),
		ClassroomID: utils.PtrValue(session.ClassroomID),
		SessionType: session.SessionType,
		TargetID:    session.TargetID,
		StartTime:   session.StartTime,
		EndTime:     session.EndTime,
	}, nil
}

// 关闭会话，同时需要从 message 表中查询所有参与会话的对象，并更新到 participants 字段
func (d *BehaviorDAOImpl) CloseCommunicationSession(ctx context.Context, sessionID string) error {
	records, err := d.communicationSessionDao.GetSessionsByIDs(ctx, []string{sessionID})
	if err != nil {
		return errors.Wrap(err, "find communication session failed")
	}
	if len(records) == 0 {
		return errors.New("communication session not found")
	}
	communicationSession := records[0]

	// 会话已关闭，不能再操作
	if communicationSession.Closed {
		return nil
	}

	communicationMessages, err := d.communicationMessageDao.GetSessionParticipants(ctx, sessionID, nil)
	if err != nil {
		return errors.Wrap(err, "find communication messages failed")
	}

	participants := make(map[string][]uint64, 0)
	for _, message := range communicationMessages {
		if _, ok := participants[message.UserType]; !ok {
			participants[message.UserType] = make([]uint64, 0)
		}
		participants[message.UserType] = append(participants[message.UserType], message.UserID)
	}

	// 转成 json string 存储 ch
	participantsJSON, err := json.Marshal(participants)
	if err != nil {
		return errors.Wrap(err, "marshal participants failed")
	}

	if err := d.communicationSessionDao.UpdateCommunicationSession(ctx,
		map[string]any{"end_time": time.Now(), "participants": participantsJSON, "closed": true},
		map[string]any{"session_id": sessionID}); err != nil {
		return errors.Wrap(err, "update communication session failed")
	}

	return nil
}

// SaveTeacherBehavior 保存教师行为
func (d *BehaviorDAOImpl) SaveTeacherBehavior(ctx context.Context, behaviors []*dto.TeacherBehaviorDTO) error {
	// 转换为数据库模型
	dbModels := make([]*TeacherBehavior, len(behaviors))
	for i, b := range behaviors {
		dbModels[i] = &TeacherBehavior{
			SchoolID:               b.SchoolID,
			ClassID:                b.ClassID,
			ClassroomID:            b.ClassroomID,
			TeacherID:              b.TeacherID,
			BehaviorType:           string(b.BehaviorType),
			CommunicationSessionID: b.CommunicationSessionID,
			Context:                b.Context,
			CreateTime:             b.CreateTime,
			UpdateTime:             time.Now(),
		}
	}

	// 批量保存
	if err := d.teacherBehaviorDao.SaveTeacherBehaviors(ctx, dbModels); err != nil {
		return errors.Wrap(err, "save teacher behavior failed")
	}

	return nil
}

// SaveStudentBehavior 保存学生行为
func (d *BehaviorDAOImpl) SaveStudentBehavior(ctx context.Context, behaviors []*dto.StudentBehaviorDTO) error {
	// 转换为数据库模型
	dbModels := make([]*StudentBehavior, len(behaviors))
	for i, b := range behaviors {
		dbModels[i] = &StudentBehavior{
			SchoolID:               b.SchoolID,
			ClassID:                b.ClassID,
			ClassroomID:            b.ClassroomID,
			StudentID:              b.StudentID,
			BehaviorType:           string(b.BehaviorType),
			CommunicationSessionID: b.CommunicationSessionID,
			Context:                b.Context,
			CreateTime:             b.CreateTime,
			UpdateTime:             time.Now(),
		}
	}

	// 批量保存
	if err := d.studentBehaviorDao.SaveStudentBehaviors(ctx, dbModels); err != nil {
		return errors.Wrap(err, "save student behavior failed")
	}

	return nil
}

// SaveCommunication 保存沟通记录
func (d *BehaviorDAOImpl) SaveCommunication(ctx context.Context, sessions []*dto.CommunicationSessionDTO, messages []*dto.CommunicationMessageDTO) error {
	// 保存会话
	if len(sessions) > 0 {
		dbSessions := make([]*CommunicationSession, len(sessions))
		for i, s := range sessions {
			dbSessions[i] = &CommunicationSession{
				SessionID:   s.SessionID,
				UserID:      s.UserID,
				UserType:    s.UserType,
				SchoolID:    s.SchoolID,
				CourseID:    utils.Ptr(s.CourseID),
				ClassroomID: utils.Ptr(s.ClassroomID),
				SessionType: s.SessionType,
				TargetID:    s.TargetID,
				StartTime:   s.StartTime,
				EndTime:     s.EndTime,
			}
		}
		if err := d.communicationSessionDao.SaveCommunicationSessions(ctx, dbSessions); err != nil {
			return errors.Wrap(err, "save communication sessions failed")
		}
	}

	if len(messages) == 0 {
		return nil
	}

	// 检查消息中的会话 id 是否存在
	sessionIds := make([]string, 0)
	for _, m := range messages {
		sessionIds = append(sessionIds, m.SessionID)
	}
	exist, err := d.communicationSessionDao.CheckSessionIDsExist(ctx, sessionIds)
	if err != nil {
		return errors.Wrap(err, "check communication sessions failed")
	}
	if !exist {
		return errors.New("session not found")
	}

	// 检查消息 id 是否存在
	answerMsgIds := make([]string, 0)
	for _, m := range messages {
		if m.AnswerTo != "" {
			answerMsgIds = append(answerMsgIds, m.AnswerTo)
		}
	}
	exist, err = d.communicationMessageDao.CheckMessageIDsExist(ctx, answerMsgIds)
	if err != nil {
		return errors.Wrap(err, "check communication messages failed")
	}
	if !exist {
		return errors.New("message not found")
	}

	dbModels := make([]*CommunicationMessage, 0, len(messages))
	for _, m := range messages {
		dbModels = append(dbModels, &CommunicationMessage{
			MessageID:      m.MessageID,
			SessionID:      m.SessionID,
			UserID:         m.UserID,
			UserType:       m.UserType,
			MessageContent: m.MessageContent,
			MessageType:    m.MessageType,
			AnswerTo:       m.AnswerTo,
			CreatedAt:      m.CreatedAt,
		})
	}
	if err := d.communicationMessageDao.SaveCommunicationMessages(ctx, dbModels); err != nil {
		return errors.Wrap(err, "save communication message failed")
	}

	return nil
}

// SaveCommunicationSession 保存单个沟通会话
func (d *BehaviorDAOImpl) SaveCommunicationSession(ctx context.Context, session *dto.CommunicationSessionDTO) error {
	sessionModel := &CommunicationSession{
		SessionID:   session.SessionID,
		UserID:      session.UserID,
		UserType:    session.UserType,
		SchoolID:    session.SchoolID,
		CourseID:    utils.Ptr(session.CourseID),
		ClassroomID: utils.Ptr(session.ClassroomID),
		SessionType: session.SessionType,
		TargetID:    session.TargetID,
		StartTime:   session.StartTime,
		EndTime:     session.EndTime,
	}
	return d.communicationSessionDao.SaveCommunicationSessions(ctx, []*CommunicationSession{sessionModel})
}

// SaveCommunicationMessage 保存单个沟通消息
func (d *BehaviorDAOImpl) SaveCommunicationMessage(ctx context.Context, message *dto.CommunicationMessageDTO) error {
	messageModel := &CommunicationMessage{
		MessageID:      message.MessageID,
		SessionID:      message.SessionID,
		UserID:         message.UserID,
		UserType:       message.UserType,
		MessageContent: message.MessageContent,
		MessageType:    message.MessageType,
		AnswerTo:       message.AnswerTo,
		CreatedAt:      message.CreatedAt,
	}
	return d.communicationMessageDao.SaveCommunicationMessages(ctx, []*CommunicationMessage{messageModel})
}

// BatchSave 批量保存数据
func (d *BehaviorDAOImpl) BatchSave(ctx context.Context, behaviors []any) error {
	// 按类型分组
	teacherBehaviors := make([]*dto.TeacherBehaviorDTO, 0)
	studentBehaviors := make([]*dto.StudentBehaviorDTO, 0)
	sessions := make([]*dto.CommunicationSessionDTO, 0)
	messages := make([]*dto.CommunicationMessageDTO, 0)

	for _, b := range behaviors {
		switch v := b.(type) {
		case *dto.TeacherBehaviorDTO:
			teacherBehaviors = append(teacherBehaviors, v)
		case *dto.StudentBehaviorDTO:
			studentBehaviors = append(studentBehaviors, v)
		case *dto.CommunicationSessionDTO:
			sessions = append(sessions, v)
		case *dto.CommunicationMessageDTO:
			messages = append(messages, v)
		default:
			return fmt.Errorf("unsupported behavior type: %T", v)
		}
	}

	// 批量保存教师行为
	if len(teacherBehaviors) > 0 {
		if err := d.SaveTeacherBehavior(ctx, teacherBehaviors); err != nil {
			return err
		}
	}

	// 批量保存学生行为
	if len(studentBehaviors) > 0 {
		if err := d.SaveStudentBehavior(ctx, studentBehaviors); err != nil {
			return err
		}
	}

	// 批量保存沟通会话和消息
	if err := d.SaveCommunication(ctx, sessions, messages); err != nil {
		return err
	}

	return nil
}

// 查询指定会话的全部消息
func (d *BehaviorDAOImpl) GetCommunicationSessionMessages(ctx context.Context, sessionID string, pageInfo *consts.DBPageInfo) ([]*dto.CommunicationMessageDTO, error) {
	records, err := d.communicationMessageDao.GetSessionMessages(ctx, sessionID, pageInfo)
	if err != nil {
		return nil, errors.Wrap(err, "query communication session messages failed")
	}

	var messages []*dto.CommunicationMessageDTO
	for _, msg := range records {
		messages = append(messages, &dto.CommunicationMessageDTO{
			MessageID:      msg.MessageID,
			SessionID:      msg.SessionID,
			UserID:         msg.UserID,
			UserType:       string(msg.UserType),
			MessageContent: msg.MessageContent,
			MessageType:    msg.MessageType,
			AnswerTo:       msg.AnswerTo,
			CreatedAt:      msg.CreatedAt,
		})
	}
	return messages, nil
}

// 查询某个会话指定 id 的消息列表
func (d *BehaviorDAOImpl) GetCommunicationSessionMessagesByIDs(ctx context.Context, sessionID string, messageIDs []string) ([]*dto.CommunicationMessageDTO, error) {
	records, err := d.communicationMessageDao.GetMessagesByIDs(ctx, messageIDs)
	if err != nil {
		return nil, errors.Wrap(err, "query communication session messages failed")
	}

	var messages []*dto.CommunicationMessageDTO
	for _, msg := range records {
		messages = append(messages, &dto.CommunicationMessageDTO{
			MessageID:      msg.MessageID,
			SessionID:      msg.SessionID,
			UserID:         msg.UserID,
			UserType:       string(msg.UserType),
			MessageContent: msg.MessageContent,
			MessageType:    msg.MessageType,
			AnswerTo:       msg.AnswerTo,
			CreatedAt:      msg.CreatedAt,
		})
	}
	return messages, nil
}

// 查询指定课堂的全部消息
func (d *BehaviorDAOImpl) GetClassroomMessages(ctx context.Context, classroomID string, pageInfo *consts.DBPageInfo) ([]*dto.CommunicationMessageDTO, error) {
	records, err := d.communicationMessageDao.GetClassroomMessages(ctx, classroomID, pageInfo)
	if err != nil {
		return nil, errors.Wrap(err, "query classroom messages failed")
	}

	var messages []*dto.CommunicationMessageDTO
	for _, msg := range records {
		messages = append(messages, &dto.CommunicationMessageDTO{
			MessageID:      msg.MessageID,
			SessionID:      msg.SessionID,
			UserID:         msg.UserID,
			UserType:       string(msg.UserType),
			MessageContent: msg.MessageContent,
			MessageType:    msg.MessageType,
			AnswerTo:       msg.AnswerTo,
			CreatedAt:      msg.CreatedAt,
		})
	}
	return messages, nil
}

// GetClassLatestBehaviors 获取班级学生最新行为
func (d *BehaviorDAOImpl) GetClassLatestBehaviors(ctx context.Context, ClassroomID uint64) ([]*dto.StudentLatestBehaviorDTO, error) {
	// 调用底层DAO获取数据
	behaviors, err := d.studentBehaviorDao.GetClassLatestBehaviors(ctx, ClassroomID)
	if err != nil {
		d.logger.Error(ctx, "获取班级%d学生最新行为失败: %v", ClassroomID, err)
		return nil, err
	}

	// 转换为DTO对象
	result := make([]*dto.StudentLatestBehaviorDTO, len(behaviors))
	for i, behavior := range behaviors {
		dto := &dto.StudentLatestBehaviorDTO{
			StudentID:      behavior.StudentID,
			BehaviorType:   consts.BehaviorType(behavior.BehaviorType),
			Context:        behavior.Context,
			LastUpdateTime: behavior.LastUpdateTime,
		}

		// 尝试解析上下文JSON
		var contextMap map[string]interface{}
		if err := json.Unmarshal([]byte(behavior.Context), &contextMap); err != nil {
			d.logger.Error(ctx, "解析行为上下文失败: %v", err)
			continue
		}

		// 提取字段并记录转换失败的情况
		d.extractBehaviorFields(ctx, dto, contextMap)

		// 解析答题信息 - 仅对答题行为处理
		if dto.BehaviorType == consts.BehaviorTypeAnswer {
			parseQuestionsInfo(ctx, d.logger, dto, contextMap)
		}

		result[i] = dto
	}

	return result, nil
}

// extractBehaviorFields 提取行为字段
func (d *BehaviorDAOImpl) extractBehaviorFields(ctx context.Context, dto *dto.StudentLatestBehaviorDTO, contextMap map[string]interface{}) {
	dto.PageName = utils.GetMapValueString(contextMap, "page_name", "")
	dto.Subject = utils.GetMapValueString(contextMap, "subject", "")
	dto.LearningType = utils.GetMapValueString(contextMap, "learning_type", "")
	dto.VideoStatus = utils.GetMapValueString(contextMap, "video_status", "")
	dto.MaterialID = utils.GetMapValueUint64(contextMap, "material_id", 0)
	dto.StayDuration = utils.GetMapValueI64(contextMap, "stay_duration", 0)
}

// parseQuestionsInfo 解析答题信息辅助函数
func parseQuestionsInfo(ctx context.Context, logger *clogger.ContextLogger, dto *dto.StudentLatestBehaviorDTO, contextMap map[string]interface{}) {
	if questionsInfo, ok := contextMap["questions_info"].(map[string]interface{}); ok {
		dto.TotalQuestions = utils.GetMapValueI64(questionsInfo, "total_questions", 0)
		dto.CorrectAnswers = utils.GetMapValueI64(questionsInfo, "correct_answers", 0)
		dto.WrongAnswers = utils.GetMapValueI64(questionsInfo, "wrong_answers", 0)
		dto.AccuracyRate = utils.F64Percent(float64(dto.CorrectAnswers), float64(dto.TotalQuestions), 2)
	}
}

// GetStudentClassroomDetail 获取学生课堂详情
func (d *BehaviorDAOImpl) GetStudentClassroomDetail(ctx context.Context, studentID, classroomID uint64) (*dto.StudentClassroomDetailDTO, error) {
	return d.studentBehaviorDao.GetStudentClassroomDetail(ctx, studentID, classroomID)
}

// GetClassBehaviorCategory 获取课堂行为分类
func (d *BehaviorDAOImpl) GetClassBehaviorCategory(ctx context.Context, classID, classroomID uint64) (*dto.ClassBehaviorCategoryDTO, error) {
	// 这里直接返回空结构，实际处理逻辑已经在Handler层完成
	// 如果后续需要从数据库直接查询分类数据，可以在这里实现
	return &dto.ClassBehaviorCategoryDTO{
		ClassroomID:   classroomID,
		QueryTime:     time.Now().Unix(),
		PraiseList:    []dto.StudentBehaviorCategoryDTO{},
		AttentionList: []dto.StudentBehaviorCategoryDTO{},
		HandledList:   []dto.StudentBehaviorCategoryDTO{},
	}, nil
}

// GetStudentsBehaviors 获取指定学生ID列表的行为数据
func (d *BehaviorDAOImpl) GetStudentsBehaviors(ctx context.Context, studentIDs []uint64) ([]*dto.StudentLatestBehaviorDTO, error) {
	// 获取原始行为数据
	behaviors, err := d.studentBehaviorDao.GetStudentsBehaviors(ctx, studentIDs)
	if err != nil {
		d.logger.Error(ctx, "查询学生行为数据失败: %v", err)
		return nil, err
	}

	// 转换为DTO对象
	result := make([]*dto.StudentLatestBehaviorDTO, len(behaviors))
	for i, behavior := range behaviors {
		dto := &dto.StudentLatestBehaviorDTO{
			StudentID:      int64(behavior.StudentID),
			BehaviorType:   consts.BehaviorType(behavior.BehaviorType),
			Context:        behavior.Context,
			LastUpdateTime: behavior.CreateTime.Unix(),
		}

		// 尝试解析上下文JSON
		var contextMap map[string]interface{}
		if err := json.Unmarshal([]byte(behavior.Context), &contextMap); err != nil {
			d.logger.Error(ctx, "解析行为上下文失败: %v", err)
			continue
		}

		// 提取字段
		d.extractBehaviorFields(ctx, dto, contextMap)

		// 解析答题信息 - 仅对答题行为处理
		if dto.BehaviorType == consts.BehaviorTypeAnswer {
			parseQuestionsInfo(ctx, d.logger, dto, contextMap)
		}

		result[i] = dto
	}

	return result, nil
}

// GetClassAllBehaviors 获取班级学生所有行为（用于汇总统计）
func (d *BehaviorDAOImpl) GetClassAllBehaviors(ctx context.Context, ClassroomID uint64) ([]*dto.StudentLatestBehaviorDTO, error) {
	// 调用底层DAO获取数据
	behaviors, err := d.studentBehaviorDao.GetClassAllBehaviors(ctx, ClassroomID)
	if err != nil {
		d.logger.Error(ctx, "获取班级%d学生所有行为失败: %v", ClassroomID, err)
		return nil, err
	}

	// 转换为DTO对象
	result := make([]*dto.StudentLatestBehaviorDTO, len(behaviors))
	for i, behavior := range behaviors {
		dto := &dto.StudentLatestBehaviorDTO{
			StudentID:      behavior.StudentID,
			BehaviorType:   consts.BehaviorType(behavior.BehaviorType),
			Context:        behavior.Context,
			LastUpdateTime: behavior.LastUpdateTime,
		}

		// 尝试解析上下文JSON
		var contextMap map[string]interface{}
		if err := json.Unmarshal([]byte(behavior.Context), &contextMap); err != nil {
			d.logger.Error(ctx, "解析行为上下文失败: %v", err)
			continue
		}

		// 提取字段并记录转换失败的情况
		d.extractBehaviorFields(ctx, dto, contextMap)

		// 解析答题信息 - 仅对答题行为处理
		if dto.BehaviorType == consts.BehaviorTypeAnswer {
			parseQuestionsInfo(ctx, d.logger, dto, contextMap)
		}

		result[i] = dto
	}

	return result, nil
}

// CountStudentTaskPraiseAndAttention 统计学生任务点赞和关注
func (d *BehaviorDAOImpl) CountStudentTaskPraiseAndAttention(ctx context.Context, taskID, assignID uint64, studentIDs []uint64) ([]dao.CHGroupCountResult, error) {
	return d.teacherBehaviorDao.CountStudentTaskPraiseAndAttention(ctx, taskID, assignID, studentIDs)
}

// GetStudentBehaviorsByType 获取学生特定类型的行为数据
func (d *BehaviorDAOImpl) GetStudentBehaviorsByType(ctx context.Context, studentID, classroomID uint64, behaviorType string) ([]*StudentBehavior, error) {
	return d.studentBehaviorDao.GetStudentBehaviorsByType(ctx, studentID, classroomID, behaviorType)
}
