package itl

// 与题库平台 API 交互的请求和响应

// 获取业务树列表请求
type ListBizTreeRequestBody struct {
	// BaseTreeId     int64   `json:"baseTreeId,omitempty"`     // 教师端暂不使用
	// BizTreeId      int64   `json:"bizTreeId,omitempty"`      // 教师端暂不使用
	// BizTreeNameKey string  `json:"bizTreeNameKey,omitempty"` // 教师端暂不使用
	// MaterialList   []int64 `json:"materialList,omitempty"`   // 教师端暂不使用
	BizTreeType int64   `json:"bizTreeType"` // 业务树类型，1 章节类型，2 知识点类型
	PhaseList   []int64 `json:"phaseList,omitempty"`
	SubjectList []int64 `json:"subjectList,omitempty"`
	Page        int64   `json:"page,omitempty"`
	PageSize    int64   `json:"pageSize,omitempty"`
}

// 获取业务树列表响应
type ListBizTreeResponseBody struct {
	Code    int64               `json:"code"`
	Message string              `json:"message"`
	Data    BizTreeInfoListPage `json:"data"`
}

type BizTreeInfoListPage struct {
	BizTreeInfoList []BizTreeInfo `json:"list"`
	Page            int64         `json:"page"`
	PageSize        int64         `json:"pageSize"`
	Total           int64         `json:"total"`
}

type BizTreeInfo struct {
	BizTreeId   int64  `json:"bizTreeId"`
	BizTreeType int64  `json:"bizTreeType"`
	BizTreeName string `json:"bizTreeName"`
	// BaseTreeId     int64  `json:"baseTreeId"` // 教师端暂不使用
	BizTreeVersion string `json:"bizTreeVersion"`
	Phase          int64  `json:"phase"`
	Subject        int64  `json:"subject"`
	Material       int64  `json:"material"` // 教材版本
	// CreaterId      int64  `json:"createrId"`  // 教师端暂不使用
	// UpdaterId      int64  `json:"updaterId"`  // 教师端暂不使用
	// CreateTime     int64  `json:"createTime"` // 教师端暂不使用
	// UpdateTime     int64  `json:"updateTime"` // 教师端暂不使用
}

// 获取业务树详情请求
type GetBizTreeDetailRequestQuery struct {
	BizTreeID       int64 `json:"bizTreeId"`
	ShowShelfStatus int64 `json:"showShelfStatus,omitempty"` // 展示上架状态，默认是0，设置1则展示
}

// 获取业务树详情响应
type GetBizTreeDetailResponseBody struct {
	Code    int64         `json:"code"`
	Message string        `json:"message"`
	Data    BizTreeEntity `json:"data"`
}

type BizTreeEntity struct {
	BizTreeId      int64  `json:"bizTreeId"`      // 业务树id
	BizTreeType    int64  `json:"bizTreeType"`    // 业务树类型
	BizTreeName    string `json:"bizTreeName"`    // 业务树名称
	BizTreeVersion string `json:"bizTreeVersion"` // 业务树版本
	// Material       int64              `json:"material"`       // 教材版本枚举 // 教师端暂不使用
	// BaseTreeId     int64              `json:"baseTreeId"`     // 基础树id // 教师端暂不使用
	Phase         int64              `json:"phase"`         // 阶段枚举
	Subject       int64              `json:"subject"`       // 科目枚举
	BizTreeDetail *BizTreeNodeEntity `json:"bizTreeDetail"` // 业务树节点列表
}

type BizTreeNodeEntity struct {
	BizTreeId       int64  `json:"bizTreeId"`       // 业务树id
	BizTreeName     string `json:"bizTreeName"`     // 业务树名称
	BizTreeNodeId   int64  `json:"bizTreeNodeId"`   // 业务树节点id
	BizTreeNodeName string `json:"bizTreeNodeName"` // 业务树节点名称
	// BizTreeParentNodeId     int64                `json:"bizTreeParentNodeId"`     // 业务树父节点id // 教师端暂不使用
	// BizTreeNodeSiblingOrder int64                `json:"bizTreeNodeSiblingOrder"` // 业务树节点兄弟节点顺序 // 教师端暂不使用
	BizTreeNodeLevel int64  `json:"bizTreeNodeLevel"` // 业务树节点层级
	BizTreeDetail    string `json:"bizTreeDetail"`    // 业务树详情
	// BaseTreeNodeIDs        []int64              `json:"baseTreeNodeIds"`        // 关联基础树节点id列表 关联知识点 // 教师端暂不使用
	// ShelfStatus            int64                `json:"shelfStatus"`            // 叶子上架状态 // 教师端暂不使用
	// NoLeafNodeShelfOnCount int64                `json:"noLeafNodeShelfOnCount"` // 非叶子节点上架数量 // 教师端暂不使用
	// NoLeafNodeTotalCount   int64                `json:"noLeafNodeTotalCount"`   // 非叶子节点全部数量 // 教师端暂不使用
	// ShelfOnLeafNodeStat    string               `json:"shelfOnLeafNodeStat"`    // 非叶子节点上架/全部 // 教师端暂不使用
	// BaseTreeNodeNameMap    map[int64]string     `json:"baseTreeNodeNameMap"`    // 关联基础树节点名称map 关联知识点 => 基础树节点名称 // 教师端暂不使用
	BizTreeNodeChildren []*BizTreeNodeEntity `json:"bizTreeNodeChildren"` // 子节点列表
}

// 业务树叶子节点，查询课程和巩固练习列表使用
type BizTreeLeafNode struct {
	BizTreeNodeId   int64  `json:"bizTreeNodeId"`   // 业务树节点id
	BizTreeNodeName string `json:"bizTreeNodeName"` // 业务树节点名称
}

// 根据id检查题集列表请求，该接口没有返回题目数据
type CheckQuestionSetExistByIDsRequestBody struct {
	QuestionSetIds []int64 `json:"questionSetIds"`
}

// 根据id检查题集列表响应，该接口没有返回题目数据
type CheckQuestionSetExistByIDsResponseBody struct {
	Code    int64  `json:"code"`
	Message string `json:"message"`
	Data    struct {
		QuestionSetList []QuestionSetShelfInfo `json:"list"`
	} `json:"data"`
}

type QuestionSetShelfInfo struct {
	QuestionSetId int64 `json:"questionSetId"`
	SceneCategory int64 `json:"sceneCategory"`
	BizTreeNodeId int64 `json:"bizTreeNodeId"`
	ShelfStatus   int64 `json:"shelfStatus"`
}

// 根据业务树叶子节点或者题集ID查询题集的响应
type GetQustionSetInfoResponseBody struct {
	Code    int64                 `json:"code"`
	Message string                `json:"message"`
	Data    QuestionSetStableInfo `json:"data"`
}

type QuestionSetStableInfo struct {
	// QuestionSetDifficultStat    []*DifficultStat           `json:"difficultStat"` // 教师端暂不使用
	QuestionSetId     int64 `json:"questionSetId"`
	QuestionVersionId int64 `json:"questionSetVersionId"`
	// AuditTaskId                 int64                      `json:"auditTaskId"` // 教师端暂不使用
	QuestionGroupStableInfoList []*QuestionGroupStableInfo `json:"list"`
}
type QuestionGroupStableInfo struct {
	QuestionGroupId    int64                 `json:"questionGroupId"`
	QuestionGroupName  string                `json:"questionGroupName"`
	QuestionGroupOrder int64                 `json:"questionGroupOrder"`
	QuestionInfoList   []*QuestionStableInfo `json:"questionGroupQuestionList"`
}

type QuestionStableInfo struct {
	QuestionId   string    `json:"questionId"`
	QuestionInfo *Question `json:"questionInfo"`
}

type DifficultStat struct {
	DifficultName              string `json:"difficultName"`
	DifficultNeedQuestionCount int64  `json:"difficultNeedQuestionCount"`
	DifficultQuestionCount     int64  `json:"difficultQuestionCount"`
}

// 前端目前使用的结构体
type Question struct {
	QuestionVersionId int64  `json:"questionVersionId"`
	QuestionId        string `json:"questionId"`
	*QuestionInfoEntity
	*QuestionContentEntity
	*QuestionProfileEntity
	AttachFileList  []*AttachFile `json:"attachFileList"`
	SubQuestionList []*Question   `json:"subQuestionList"`
	QuestionTags    []string      `json:"questionTags"` // 教师端额外添加的字段，题目标签
	// CreaterId    int64    `json:"createrId"`              // 教师端暂不使用
	// UpdaterId    int64    `json:"updaterId"`              // 教师端暂不使用
	// CreateTime   int64    `json:"createTime"`             // 教师端暂不使用
	// UpdateTime   int64    `json:"updateTime"`             // 教师端暂不使用
}

type QuestionInfoEntity struct {
	QuestionType       int64   `json:"questionType"`
	QuestionAnswerMode int64   `json:"questionAnswerMode"`
	IsNewestVersion    int64   `json:"isNewestVersion"`
	IsNewestPublish    int64   `json:"isNewestPublish"`
	IsPublished        int64   `json:"isPublished"`
	AuditTaskId        int64   `json:"auditTaskId"`
	QuestionYear       int64   `json:"questionYear"`
	QuestionTopic      string  `json:"questionTopic"`
	QuestionDifficult  int64   `json:"questionDifficult"`
	BaseTreeId         int64   `json:"baseTreeId"`
	BaseTreeNodeIds    []int64 `json:"baseTreeNodeIds"`
	ProvinceCode       int64   `json:"provinceCode"`
	CityCode           int64   `json:"cityCode"`
	AreaCode           int64   `json:"areaCode"`
	Phase              int64   `json:"phase"`
	Subject            int64   `json:"subject"`
	QuestionMd5        string  `json:"questionMd5"`
	EstimatedDuration  int64   `json:"estimatedDuration"`
}

type QuestionContent struct {
	QuestionId string `json:"questionId"`
	*QuestionContentEntity
	CreaterId  int64 `json:"-"`
	UpdaterId  int64 `json:"-"`
	CreateTime int64 `json:"-"`
	UpdateTime int64 `json:"-"`
}

type QuestionContentEntity struct {
	QuestionContentFormat *QuestionContentFormat `json:"questionContent"`
	QuestionAnswer        *QuestionAnswerFormat  `json:"questionAnswer"`
	QuestionExplanation   string                 `json:"questionExplanation"`
	QuestionExtra         string                 `json:"questionExtra"`
	CheckInfo             *CheckInfo             `json:"checkInfo"`
}

type QuestionProfileEntity struct {
	QuestionSceneList []int64 `json:"questionSceneList"`
	QuestionSource    int64   `json:"questionSource"`
	OriginPaperId     string  `json:"originPaperId"`
}

type AttachFile struct {
	FileName  string `json:"fileName"`
	FileType  string `json:"fileType"`
	FileData  string `json:"fileData"`
	OssBucket string `json:"ossBucket"`
	OssPath   string `json:"ossPath"`
}

type QuestionAnswerFormat struct {
	AnswerOptionList []*QuestionOption `json:"answerOptionList"` // 题目答案选项列表，比如 选择题的答案选项列表
}

type QuestionOption struct {
	OptionKey string `json:"optionKey"` // 选项key
	OptionVal string `json:"optionVal"` // 选项值
}

// 题目内容格式
type QuestionContentFormat struct {
	QuestionOrder      int64             `json:"questionOrder"`      //题目对应试卷的题号
	QuestionScore      float64           `json:"questionScore"`      //题目对应试卷的分数
	QuestionOriginName string            `json:"questionOriginName"` //题目来源名称 如 2022秋•房山区期末
	QuestionStem       string            `json:"questionStem"`       // 题干部分，选择题不包含选项, 填空题需要使用<blank/>区分空
	QuestionOptionList []*QuestionOption `json:"questionOptionList"` // 题目选项
}

type InvalidField struct {
	Field string `json:"field"`
	Value string `json:"value"`
}

type CheckInfo struct {
	InvalidFieldList []*InvalidField `json:"invalidFieldList"`
}

// 获取查询题目支持的全部筛选项
type QuestionEnumsResponseBody struct {
	Code    int64             `json:"code"`
	Message string            `json:"message"`
	Data    QuestionEnumsData `json:"data"`
}

type QuestionEnumsData struct {
	AuditTypeList         []EnumItem     `json:"-"`
	AuditsStatusList      []EnumItem     `json:"-"`
	MaterialList          []EnumItem     `json:"-"`
	PaperTypeList         []EnumItem     `json:"-"`
	PhaseList             []EnumItem     `json:"-"`
	SubjectList           []EnumItem     `json:"-"`
	PhaseSubjectRelation  []PhaseSubject `json:"-"`
	ProvinceList          []EnumItem     `json:"provinceList"`
	QuestionDifficultList []EnumItem     `json:"questionDifficultList"`
	QuestionSourceList    []EnumItem     `json:"-"`
	QuestionTypeList      []EnumItem     `json:"questionTypeList"`
	YearList              []EnumItem     `json:"yearList"`
	SceneCategoryList     []EnumItem     `json:"-"`
	ShelfStatusList       []EnumItem     `json:"-"`
}

type EnumItem struct {
	NameEn string `json:"nameEn"`
	NameZh string `json:"nameZh"`
	Value  int64  `json:"value"`
}
type PhaseSubject struct {
	NameEn      string     `json:"nameEn"`
	NameZh      string     `json:"nameZh"`
	Value       int64      `json:"value"`
	SubjectList []EnumItem `json:"subjectList"`
}

// 查询题目列表请求
type QuestionListRequestBody struct {
	PhaseList   []int64 `json:"phaseList"`
	SubjectList []int64 `json:"subjectList"`
	// BaseTreeNodeIds   []int64 `json:"baseTreeNodeIds,omitempty"` // 教师端暂不使用
	BizTreeNodeIds    []int64 `json:"bizTreeNodeIds,omitempty"`
	Keyword           string  `json:"keyword,omitempty"`
	QuestionType      []int64 `json:"questionType,omitempty"`
	QuestionDifficult []int64 `json:"questionDifficult,omitempty"`
	QuestionYears     []int64 `json:"questionYears,omitempty"`
	// QuestionSource    []int64 `json:"questionSource,omitempty"` // 教师端暂不使用
	// SceneCategory     []int64 `json:"sceneCategory,omitempty"`  // 教师端暂不使用
	Page     int64  `json:"page,omitempty"`
	PageSize int64  `json:"pageSize,omitempty"`
	Sort     string `json:"sort,omitempty"` // 题库支持：createTime 最新题目，useCount 最多使用
}

// 查询题目列表响应
type QuestionListResponseBody struct {
	Code    int64              `json:"code"`
	Message string             `json:"message"`
	Data    QuestionListOutput `json:"data"`
}

type QuestionListOutput struct {
	Total     int64       `json:"total"`
	Page      int64       `json:"page"`
	PageSize  int64       `json:"pageSize"`
	Questions []*Question `json:"list"`
}

// 获取题目详情请求
type GetQuestionDetailRequestQuery struct {
	QuestionId string `json:"questionId"`
}

// 获取题目详情响应
type GetQuestionDetailResponseBody struct {
	Code    int64     `json:"code"`
	Message string    `json:"message"`
	Data    *Question `json:"data"`
}

// 根据id获取题目列表请求
type GetQuestionListByIDRequestBody struct {
	QuestionIdList []string `json:"questionIdList"` // 题目id列表，最多100个
	NeedContent    int64    `json:"needContent"`    // 是否需要题目内容，0 不需要，1 需要
}

// 根据id获取题目列表响应
type GetQuestionListByIDResponseBody struct {
	Code    int64  `json:"code"`
	Message string `json:"message"`
	Data    struct {
		QuestionList []*Question `json:"list"`
	} `json:"data"`
}

// 题目精简信息，只给必要信息，避免数据泄露 - 暂未使用
type QuestionSimpleInfo struct {
	QuestionID    string `json:"questionId"`    // 题目ID
	QuestionType  int64  `json:"questionType"`  // 题目类型
	Content       string `json:"content"`       // 题目内容
	Answer        string `json:"answer"`        // 答案
	AnswerExplain string `json:"answerExplain"` // 答案解析
}
