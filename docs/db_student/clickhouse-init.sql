-- =============================================
-- 学生行为记录表
-- =============================================
CREATE TABLE db_student.tbl_student_behavior_logs
(
    id                       UUID     DEFAULT generateUUIDv4() COMMENT '主键ID',
    school_id               UInt64                            COMMENT '学校ID',
    class_id                UInt64                            COMMENT '班级ID',
    classroom_id            UInt64                            COMMENT '课堂ID（可为空）',
    student_id              UInt64                            COMMENT '学生ID',
    behavior_type           String                            COMMENT '行为类型：浏览/表扬/提醒/会话/反馈等',
    communication_session_id String   DEFAULT ''              COMMENT '会话ID（关联sessions表主键）',
    last_message_id         String   DEFAULT ''              COMMENT '最后消息ID（关联messages表主键）',
    context                 String                            COMMENT '行为内容',
    create_time             DateTime                          COMMENT '创建时间',
    update_time             DateTime DEFAULT now()            COMMENT '更新时间（默认当前时间）'
)
ENGINE = ReplacingMergeTree(update_time)
PRIMARY KEY (school_id, class_id, student_id)
ORDER BY (school_id, class_id, student_id, communication_session_id, create_time)
SETTINGS index_granularity = 8192;