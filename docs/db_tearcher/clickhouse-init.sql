-- =============================================
-- 教师行为记录表
-- =============================================
CREATE TABLE db_teacher.tbl_teacher_behavior_logs
(
    id                       UUID    DEFAULT generateUUIDv4() COMMENT '主键ID',
    school_id               UInt64                           COMMENT '学校ID',
    class_id                UInt64                           COMMENT '班级ID',
    classroom_id            UInt64   DEFAULT 0               COMMENT '课堂ID，可以为空',
    teacher_id              UInt64                           COMMENT '教师ID',
    behavior_type           String                           COMMENT '行为类型：浏览/表扬/提醒/会话/反馈等',
    communication_session_id String   DEFAULT ''             COMMENT '会话ID，可以为空',
    last_message_id         String   DEFAULT ''             COMMENT '会话中最新已读消息ID',
    context                 String                           COMMENT '行为内容，不存储会话内容',
    task_id                 UInt64   DEFAULT 0             COMMENT '任务ID',
    assign_id               UInt64   DEFAULT 0             COMMENT '任务分配ID',
    student_id              UInt64   DEFAULT 0             COMMENT '学生ID',
    create_time             DateTime                         COMMENT '创建时间',
    update_time             DateTime                         COMMENT '更新时间'
)
ENGINE = ReplacingMergeTree(update_time)
PRIMARY KEY (school_id, class_id, teacher_id)
ORDER BY (school_id, class_id, teacher_id, classroom_id, communication_session_id, create_time)
SETTINGS index_granularity = 8192;

-- -- =============================================
-- -- 课堂学情统计表
-- -- =============================================
-- CREATE TABLE tbl_classroom_learning_stats (
--     id UUID,
--     school_id UInt64,
--     class_id UInt64,
--     course_id UInt64,
--     classroom_id UUID,
--     summary JSONB,
--     report_time Datetime
-- ) ENGINE = MergeTree()
-- PARTITION BY toYYYYMM(report_time)
-- ORDER BY (school_id, classroom_id, report_time);

-- -- 表注释：课堂学情统计表，统计课堂学习情况相关数据
-- COMMENT ON TABLE tbl_classroom_learning_stats IS '课堂学情统计表，用于汇总课堂内学生学习情况、进度等统计信息，支持教学分析';
-- -- 字段注释
-- COMMENT ON COLUMN tbl_classroom_learning_stats.id IS '统计记录唯一标识UUID';
-- COMMENT ON COLUMN tbl_classroom_learning_stats.school_id IS '学校ID';
-- COMMENT ON COLUMN tbl_classroom_learning_stats.class_id IS '班级ID';
-- COMMENT ON COLUMN tbl_classroom_learning_stats.course_id IS '课程ID';
-- COMMENT ON COLUMN tbl_classroom_learning_stats.classroom_id IS '课堂ID';
-- COMMENT ON COLUMN tbl_classroom_learning_stats.summary IS '学情统计内容（JSONB格式，包含学习数据等）';
-- COMMENT ON COLUMN tbl_classroom_learning_stats.report_time IS '统计上报时间';

-- -- =============================================
-- -- 课堂统计报表表
-- -- =============================================
-- CREATE TABLE tbl_classroom_report (
--     id UUID,
--     school_id UInt64,
--     class_id UInt64,
--     classroom_id UUID,
--     report_content String,
--     create_time Datetime,
--     update_time Datetime
-- ) ENGINE = MergeTree()
-- PARTITION BY toYYYYMM(create_time)
-- ORDER BY (school_id, classroom_id, create_time);

-- -- 表注释：课堂统计报表表，存储课堂相关统计报告内容
-- COMMENT ON TABLE tbl_classroom_report IS '课堂统计报表表，用于存储课堂教学过程中生成的统计报告，如教学质量分析报告等';
-- -- 字段注释
-- COMMENT ON COLUMN tbl_classroom_report.id IS '报告记录唯一标识UUID';
-- COMMENT ON COLUMN tbl_classroom_report.school_id IS '学校ID';
-- COMMENT ON COLUMN tbl_classroom_report.class_id IS '班级ID';
-- COMMENT ON COLUMN tbl_classroom_report.classroom_id IS '课堂ID';
-- COMMENT ON COLUMN tbl_classroom_report.report_content IS '报告内容文本';
-- COMMENT ON COLUMN tbl_classroom_report.create_time IS '报告创建时间';
-- COMMENT ON COLUMN tbl_classroom_report.update_time IS '报告更新时间';

-- =============================================
-- 沟通记录主表
-- =============================================
CREATE TABLE db_teacher.tbl_communication_sessions
(
    session_id    UUID    DEFAULT generateUUIDv4() COMMENT '会话ID',
    user_id       UInt64                           COMMENT '发起人ID',
    user_type     Enum8('student' = 1, 'teacher' = 2, 'ai' = 3) COMMENT '发起人类别',
    school_id     UInt64                           COMMENT '学校ID',
    course_id     UInt64                           COMMENT '课程ID',
    classroom_id  UInt64                           COMMENT '课堂ID',
    session_type  String                           COMMENT '会话类型',
    target_id     String                           COMMENT '关联对象ID',
    closed        Bool                             COMMENT '是否关闭',
    participants  String                           COMMENT '参与者列表JSON',
    start_time    DateTime                         COMMENT '开始时间',
    end_time      Nullable(DateTime)               COMMENT '结束时间'
)
ENGINE = ReplacingMergeTree(start_time)
PRIMARY KEY (session_id, user_id)
ORDER BY (session_id, user_id, user_type, start_time)
SETTINGS index_granularity = 8192;

-- =============================================
-- 沟通记录详情表
-- =============================================
CREATE TABLE db_teacher.tbl_communication_messages
(
    message_id       UUID    DEFAULT generateUUIDv4() COMMENT '消息ID',
    session_id       UUID                            COMMENT '会话ID',
    user_id          UInt64                          COMMENT '发送人ID',
    user_type        Enum8('student' = 1, 'teacher' = 2, 'ai' = 3) COMMENT '发送人类别',
    message_content  String                          COMMENT '文本内容',
    message_type     String                          COMMENT '消息类型',
    answer_to        String  DEFAULT ''             COMMENT '回答关联的消息ID',
    created_at       DateTime                        COMMENT '创建时间'
)
ENGINE = ReplacingMergeTree
ORDER BY (message_id, session_id, user_id, user_type)
SETTINGS index_granularity = 8192;

