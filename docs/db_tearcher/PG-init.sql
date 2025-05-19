-- 创建更新时间触发器函数
CREATE OR REPLACE FUNCTION update_timestamp()
RETURNS TRIGGER AS $$
BEGIN
    NEW.update_time = EXTRACT(EPOCH FROM CURRENT_TIMESTAMP)::BIGINT;
    RETURN NEW;
END;
$$ language 'plpgsql';


-- =============================================
-- 权限管理模块
-- =============================================

-- -- --------------------------------
-- -- 群组表
-- -- --------------------------------
-- BEGIN;
-- CREATE TABLE tbl_group (
--     permission_id VARCHAR(20) PRIMARY KEY,
--     school_id BIGINT NOT NULL,
--     group_name VARCHAR(200) NOT NULL,
--     group_description TEXT,
--     create_time BIGINT DEFAULT EXTRACT(EPOCH FROM CURRENT_TIMESTAMP)::BIGINT,
--     update_time BIGINT DEFAULT EXTRACT(EPOCH FROM CURRENT_TIMESTAMP)::BIGINT
-- );
-- -- 表注释：群组信息表，存储群组权限相关信息
-- COMMENT ON TABLE tbl_group IS '群组信息表，存储群组ID、所属学校、群组名称、描述等信息';
-- -- 字段注释
-- COMMENT ON COLUMN tbl_group.permission_id IS '群组权限ID，主键';
-- COMMENT ON COLUMN tbl_group.school_id IS '群组所属学校ID';
-- COMMENT ON COLUMN tbl_group.group_name IS '群组名称';
-- COMMENT ON COLUMN tbl_group.group_description IS '群组描述信息';
-- COMMENT ON COLUMN tbl_group.create_time IS '群组创建时间';
-- COMMENT ON COLUMN tbl_group.update_time IS '群组信息更新时间';
-- -- 创建索引
-- CREATE INDEX idx_tbl_group_school_id ON tbl_group(school_id);
-- -- 创建更新时间触发器
-- CREATE TRIGGER update_tbl_group_timestamp
--     BEFORE UPDATE ON tbl_group
--     FOR EACH ROW
--     EXECUTE FUNCTION update_timestamp();
-- COMMIT;


-- =============================================
-- 课堂管理模块
-- =============================================

-- --------------------------------
-- 课堂表
-- --------------------------------
BEGIN;
CREATE TABLE tbl_classroom (
    id BIGSERIAL PRIMARY KEY,
    teacher_id VARCHAR(20) NOT NULL,
    school_id BIGINT NOT NULL,
    class_id BIGINT NOT NULL,
    course_id BIGINT NOT NULL,
    status BIGINT NOT NULL,
    create_time BIGINT DEFAULT EXTRACT(EPOCH FROM CURRENT_TIMESTAMP)::BIGINT,
    update_time BIGINT DEFAULT EXTRACT(EPOCH FROM CURRENT_TIMESTAMP)::BIGINT
);
-- 表注释：课堂信息表，记录课堂相关数据
COMMENT ON TABLE tbl_classroom IS '课堂信息表，存储课堂ID、关联教师、学校、班级、课程、状态等信息';
-- 字段注释
COMMENT ON COLUMN tbl_classroom.id IS '课堂自增主键ID';
COMMENT ON COLUMN tbl_classroom.teacher_id IS '授课教师ID';
COMMENT ON COLUMN tbl_classroom.school_id IS '课堂所属学校ID';
COMMENT ON COLUMN tbl_classroom.class_id IS '课堂关联班级ID';
COMMENT ON COLUMN tbl_classroom.course_id IS '课程ID';
COMMENT ON COLUMN tbl_classroom.status IS '课堂状态（未开始/进行中/已结束等）';
COMMENT ON COLUMN tbl_classroom.create_time IS '课堂创建时间';
COMMENT ON COLUMN tbl_classroom.update_time IS '课堂信息更新时间';

-- 创建索引
CREATE INDEX idx_tbl_classroom_teacher_id ON tbl_classroom(teacher_id);
CREATE INDEX idx_tbl_classroom_school_id ON tbl_classroom(school_id);
-- 创建更新时间触发器
CREATE TRIGGER update_tbl_classroom_timestamp
    BEFORE UPDATE ON tbl_classroom
    FOR EACH ROW
    EXECUTE FUNCTION update_timestamp();
COMMIT;

-- --------------------------------
-- 课堂反馈表
-- --------------------------------
BEGIN;
CREATE TABLE tbl_classroom_feedback (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    school_id BIGINT NOT NULL,
    class_id BIGINT,
    content TEXT,
    remark BIGINT,
    create_type BIGINT NOT NULL,
    create_time BIGINT DEFAULT EXTRACT(EPOCH FROM CURRENT_TIMESTAMP)::BIGINT,
    update_time BIGINT DEFAULT EXTRACT(EPOCH FROM CURRENT_TIMESTAMP)::BIGINT
);
-- 表注释：课堂反馈表，记录课堂反馈信息
COMMENT ON TABLE tbl_classroom_feedback IS '课堂反馈表，存储反馈ID、用户ID、学校、班级、反馈内容等信息';
-- 字段注释
COMMENT ON COLUMN tbl_classroom_feedback.id IS '反馈自增主键ID';
COMMENT ON COLUMN tbl_classroom_feedback.user_id IS '反馈用户ID';
COMMENT ON COLUMN tbl_classroom_feedback.school_id IS '反馈所属学校ID';
COMMENT ON COLUMN tbl_classroom_feedback.class_id IS '反馈关联班级ID';
COMMENT ON COLUMN tbl_classroom_feedback.content IS '反馈内容';
COMMENT ON COLUMN tbl_classroom_feedback.remark IS '反馈备注标识';
COMMENT ON COLUMN tbl_classroom_feedback.create_type IS '反馈创建类型';
COMMENT ON COLUMN tbl_classroom_feedback.create_time IS '反馈创建时间';
COMMENT ON COLUMN tbl_classroom_feedback.update_time IS '反馈信息更新时间';

-- 创建索引
CREATE INDEX idx_tbl_classroom_feedback_user_id ON tbl_classroom_feedback(user_id);
CREATE INDEX idx_tbl_classroom_feedback_school_id ON tbl_classroom_feedback(school_id);
-- 创建更新时间触发器
CREATE TRIGGER update_tbl_classroom_feedback_timestamp
    BEFORE UPDATE ON tbl_classroom_feedback
    FOR EACH ROW
    EXECUTE FUNCTION update_timestamp();
COMMIT;

-- =============================================
-- 资源管理模块
-- =============================================

-- --------------------------------
-- 教师资源收藏表
-- --------------------------------
BEGIN;
CREATE TABLE tbl_teacher_resource_favorite (
    id BIGSERIAL PRIMARY KEY,
    teacher_id BIGINT NOT NULL,
    school_id BIGINT NOT NULL,
    resource_id VARCHAR(16) NOT NULL,
    resource_type BIGINT NOT NULL DEFAULT 0,
    status BIGINT NOT NULL DEFAULT 1,
    create_time BIGINT DEFAULT EXTRACT(EPOCH FROM CURRENT_TIMESTAMP)::BIGINT,
    update_time BIGINT DEFAULT EXTRACT(EPOCH FROM CURRENT_TIMESTAMP)::BIGINT
);
-- 表注释：教师资源收藏表，存储教师收藏的资源信息
COMMENT ON TABLE tbl_teacher_resource_favorite IS '教师资源收藏表，存储教师ID、资源ID、资源类型等信息';
-- 字段注释
COMMENT ON COLUMN tbl_teacher_resource_favorite.id IS '自增主键ID';
COMMENT ON COLUMN tbl_teacher_resource_favorite.teacher_id IS '教师ID';
COMMENT ON COLUMN tbl_teacher_resource_favorite.school_id IS '学校ID';
COMMENT ON COLUMN tbl_teacher_resource_favorite.resource_id IS '资源ID';
COMMENT ON COLUMN tbl_teacher_resource_favorite.resource_type IS '资源类型（0:默认类型）';
COMMENT ON COLUMN tbl_teacher_resource_favorite.status IS '收藏状态（1:已收藏 0:已取消）';
COMMENT ON COLUMN tbl_teacher_resource_favorite.create_time IS '创建时间';
COMMENT ON COLUMN tbl_teacher_resource_favorite.update_time IS '更新时间';

-- 创建索引
CREATE INDEX idx_tbl_teacher_resource_favorite_teacher_id ON tbl_teacher_resource_favorite(teacher_id);
CREATE INDEX idx_tbl_teacher_resource_favorite_resource_id ON tbl_teacher_resource_favorite(resource_id);
-- 创建更新时间触发器
CREATE TRIGGER update_tbl_teacher_resource_favorite_timestamp
    BEFORE UPDATE ON tbl_teacher_resource_favorite
    FOR EACH ROW
    EXECUTE FUNCTION update_timestamp();
COMMIT;

-- --------------------------------
-- 教师上传资源表
-- --------------------------------
BEGIN;
CREATE TABLE public.tbl_resource (
    resource_id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    school_id BIGINT NOT NULL,
    file_name VARCHAR(256),
    file_type BIGINT,
    status BIGINT,
    file_byte_size BIGINT,
    metadata JSONB DEFAULT '{}',
    file_hash VARCHAR(256) DEFAULT '',
    oss_bucket VARCHAR(100) DEFAULT '',
    oss_path VARCHAR(255),
    file_scope BIGINT NOT NULL,
    create_time BIGINT DEFAULT EXTRACT(EPOCH FROM CURRENT_TIMESTAMP)::BIGINT,
    update_time BIGINT DEFAULT EXTRACT(EPOCH FROM CURRENT_TIMESTAMP)::BIGINT
);
-- 创建索引
CREATE INDEX idx_tbl_resource_access_level ON public.tbl_resource USING btree (file_scope ASC NULLS LAST);
CREATE INDEX idx_tbl_resource_file_hash ON public.tbl_resource USING btree (file_hash ASC NULLS LAST);
CREATE INDEX idx_tbl_resource_file_name ON public.tbl_resource USING btree (file_name ASC NULLS LAST);
CREATE INDEX idx_tbl_resource_metadata ON public.tbl_resource USING gin (metadata);
CREATE INDEX idx_tbl_resource_oss_bucket ON public.tbl_resource USING btree (oss_bucket ASC NULLS LAST);
CREATE INDEX idx_tbl_resource_school_id ON public.tbl_resource USING btree (school_id ASC NULLS LAST);
CREATE INDEX idx_tbl_resource_user_id ON public.tbl_resource USING btree (user_id ASC NULLS LAST);
-- 表注释：教师上传资源表，存储资源 ID、用户 ID、学校、文件名称、OSS 路径等信息
COMMENT ON TABLE public.tbl_resource IS '教师上传资源表，存储资源 ID、用户 ID、学校、上传名称、文件 ID、存储路径等信息';
-- 字段注释
COMMENT ON COLUMN public.tbl_resource.resource_id IS '自增主键 ID';
COMMENT ON COLUMN public.tbl_resource.user_id IS '上传资源的用户 ID';
COMMENT ON COLUMN public.tbl_resource.school_id IS '资源所属学校 ID';
COMMENT ON COLUMN public.tbl_resource.file_name IS '文件名称';
COMMENT ON COLUMN public.tbl_resource.file_type IS '文件类型';
COMMENT ON COLUMN public.tbl_resource.status IS '资源状态（审核中/已通过/未通过等）';
COMMENT ON COLUMN public.tbl_resource.file_byte_size IS '文件大小(字节)';
COMMENT ON COLUMN public.tbl_resource.metadata IS '元数据（JSON 格式，如文档页数、视频时长等）';
COMMENT ON COLUMN public.tbl_resource.file_hash IS '文件 MD5/SHA 哈希值';
COMMENT ON COLUMN public.tbl_resource.oss_bucket IS '存储的 bucket 桶';
COMMENT ON COLUMN public.tbl_resource.oss_path IS 'OSS 存储路径';
COMMENT ON COLUMN public.tbl_resource.file_scope IS '文件访问权限';
COMMENT ON COLUMN public.tbl_resource.create_time IS '资源上传时间';
COMMENT ON COLUMN public.tbl_resource.update_time IS '资源信息更新时间';
-- 创建更新时间触发器
CREATE TRIGGER update_tbl_resource_timestamp
    BEFORE UPDATE ON public.tbl_resource
    FOR EACH ROW
    EXECUTE FUNCTION update_timestamp();
COMMIT;

-- =============================================
-- 任务管理模块
-- =============================================

-- --------------------------------
-- 任务表
-- --------------------------------
BEGIN;
CREATE TABLE tbl_task (
    task_id BIGSERIAL PRIMARY KEY,
    school_id BIGINT NOT NULL,
    phase BIGINT NOT NULL,
    subject BIGINT NOT NULL,
    task_type BIGINT NOT NULL,
    task_sub_type BIGINT NOT NULL DEFAULT 0,
    task_name VARCHAR(32) DEFAULT '',
    teacher_comment VARCHAR(256) DEFAULT '',
    task_extra_info TEXT,
    deleted BIGINT NOT NULL DEFAULT 0,
    creator_id BIGINT NOT NULL,
    updater_id BIGINT NOT NULL,
    create_time BIGINT DEFAULT EXTRACT(EPOCH FROM CURRENT_TIMESTAMP)::BIGINT,
    update_time BIGINT DEFAULT EXTRACT(EPOCH FROM CURRENT_TIMESTAMP)::BIGINT
);
-- 表注释：任务基本信息表，存储任务相关数据
COMMENT ON TABLE tbl_task IS '任务基本信息表';
-- 字段注释
COMMENT ON COLUMN tbl_task.task_id IS '任务ID，主键';
COMMENT ON COLUMN tbl_task.school_id IS '任务所属学校ID';
COMMENT ON COLUMN tbl_task.phase IS '学段枚举值';
COMMENT ON COLUMN tbl_task.subject IS '学科枚举值';
COMMENT ON COLUMN tbl_task.task_type IS '任务类型';
COMMENT ON COLUMN tbl_task.task_sub_type IS '任务子类型';
COMMENT ON COLUMN tbl_task.task_name IS '任务名称';
COMMENT ON COLUMN tbl_task.teacher_comment IS '老师留言';
COMMENT ON COLUMN tbl_task.task_extra_info IS '任务额外信息';
COMMENT ON COLUMN tbl_task.deleted IS '任务是否删除标识';
COMMENT ON COLUMN tbl_task.creator_id IS '任务创建者ID';
COMMENT ON COLUMN tbl_task.updater_id IS '任务更新者ID';
COMMENT ON COLUMN tbl_task.create_time IS '任务创建时间';
COMMENT ON COLUMN tbl_task.update_time IS '任务信息更新时间';

-- 创建索引
CREATE INDEX idx_tbl_task_creator_id ON tbl_task(creator_id, subject, task_type);
-- 创建更新时间触发器
CREATE TRIGGER update_tbl_task_timestamp
    BEFORE UPDATE ON tbl_task
    FOR EACH ROW
    EXECUTE FUNCTION update_timestamp();
COMMIT;

-- --------------------------------
-- 任务素材资源关联表
-- --------------------------------
BEGIN;
CREATE TABLE tbl_task_resource (
    id BIGSERIAL PRIMARY KEY,
    task_id BIGINT NOT NULL,
    resource_id VARCHAR(16) NOT NULL, -- 父资源ID，目前存AI课/题集（巩固练习）/试题/试卷ID
    resource_sub_ids text[], -- 子资源ID列表，当父资源为巩固练习时，这里记录巩固练习下面的题目ID
    resource_type BIGINT NOT NULL,
    resource_extra TEXT
);
-- 表注释：任务与资源关联表，记录任务关联的素材资源
COMMENT ON TABLE tbl_task_resource IS '任务与资源关联表，存储任务ID、资源ID、资源类型等关联信息';
-- 字段注释
COMMENT ON COLUMN tbl_task_resource.id IS '自增主键ID';
COMMENT ON COLUMN tbl_task_resource.task_id IS '任务ID';
COMMENT ON COLUMN tbl_task_resource.resource_id IS '资源ID';
COMMENT ON COLUMN tbl_task_resource.resource_sub_ids IS '子资源ID列表';
COMMENT ON COLUMN tbl_task_resource.resource_type IS '资源类型';
COMMENT ON COLUMN tbl_task_resource.resource_extra IS '资源额外信息';

-- 创建索引
-- CREATE INDEX idx_tbl_task_resource_resource_id ON tbl_task_resource(resource_id);

-- 创建唯一性约束
CREATE UNIQUE INDEX idx_tbl_task_resource_unique ON tbl_task_resource(task_id, resource_id, resource_type);
COMMIT;

-- --------------------------------
-- 任务与学生群组关联表
-- --------------------------------
BEGIN;
CREATE TABLE tbl_task_assign (
    assign_id BIGSERIAL PRIMARY KEY,
    task_id BIGINT NOT NULL,
    school_id BIGINT NOT NULL,
    group_type BIGINT NOT NULL,
    group_id BIGINT NOT NULL,
    start_time BIGINT NOT NULL,
    deadline BIGINT NOT NULL,
    deleted BIGINT NOT NULL DEFAULT 0
);
-- 表注释：任务分配表，记录任务与学生群组的关联关系
COMMENT ON TABLE tbl_task_assign IS '任务分配表，记录任务与学生群组的关联关系';
-- 字段注释
COMMENT ON COLUMN tbl_task_assign.assign_id IS '分配记录ID，自增主键';
COMMENT ON COLUMN tbl_task_assign.task_id IS '任务ID，关联任务表';
COMMENT ON COLUMN tbl_task_assign.school_id IS '学校ID';
COMMENT ON COLUMN tbl_task_assign.group_type IS '群组类型';
COMMENT ON COLUMN tbl_task_assign.group_id IS '群组ID，当群组类型是班级时为班级ID';
COMMENT ON COLUMN tbl_task_assign.start_time IS '任务开始时间';
COMMENT ON COLUMN tbl_task_assign.deadline IS '任务截止时间';
COMMENT ON COLUMN tbl_task_assign.deleted IS '任务分配是否删除标识，0 未删除，1 已删除';
-- 创建索引
CREATE INDEX idx_tbl_task_assign_task_id ON tbl_task_assign(task_id);
CREATE INDEX idx_tbl_task_assign_school_id_group_id ON tbl_task_assign(school_id, group_id, group_type);
COMMIT;

-- --------------------------------
-- 任务与学生关联表
-- --------------------------------
BEGIN;
CREATE TABLE tbl_task_student (
    id BIGSERIAL PRIMARY KEY,
    assign_id BIGINT NOT NULL,
    task_id BIGINT NOT NULL,
    student_id BIGINT NOT NULL
);
-- 表注释：任务与学生关联表
COMMENT ON TABLE tbl_task_student IS '任务与学生关联表';
-- 字段注释
COMMENT ON COLUMN tbl_task_student.id IS '自增主键ID';
COMMENT ON COLUMN tbl_task_student.assign_id IS '分配ID，关联任务分配表';
COMMENT ON COLUMN tbl_task_student.task_id IS '任务ID，关联任务表';
COMMENT ON COLUMN tbl_task_student.student_id IS '学生ID，关联学生表';

-- 创建索引
CREATE INDEX idx_tbl_task_student_assign_id ON tbl_task_student(assign_id);
CREATE INDEX idx_tbl_task_student_task_id ON tbl_task_student(task_id);
CREATE INDEX idx_tbl_task_student_student_id ON tbl_task_student(student_id);
COMMIT;

-- =============================================
-- 任务报告设置表
-- =============================================
-- 创建 tbl_task_report_setting 表
BEGIN;
CREATE TABLE tbl_task_report_setting (
    id BIGSERIAL PRIMARY KEY,
    school_id BIGINT NOT NULL,
    subject BIGINT NOT NULL,
    teacher_id BIGINT NOT NULL,
    class_id BIGINT NOT NULL,
    setting JSONB,
    create_time BIGINT DEFAULT EXTRACT(EPOCH FROM CURRENT_TIMESTAMP)::BIGINT,
    update_time BIGINT DEFAULT EXTRACT(EPOCH FROM CURRENT_TIMESTAMP)::BIGINT
);

-- 创建唯一性约束
CREATE UNIQUE INDEX unique_class_subject_report_setting ON tbl_task_report_setting(school_id, class_id, subject, teacher_id);

-- 表注释
COMMENT ON TABLE tbl_task_report_setting IS '任务报告设置表';
-- 字段注释
COMMENT ON COLUMN tbl_task_report_setting.id IS '自增主键，唯一标识每条记录';
COMMENT ON COLUMN tbl_task_report_setting.school_id IS '学校 ID';
COMMENT ON COLUMN tbl_task_report_setting.subject IS '学科 ID';
COMMENT ON COLUMN tbl_task_report_setting.teacher_id IS '教师 ID';
COMMENT ON COLUMN tbl_task_report_setting.class_id IS '班级 ID';
COMMENT ON COLUMN tbl_task_report_setting.setting IS '任务设置的 JSON 数据，存储具体的任务配置信息';
COMMENT ON COLUMN tbl_task_report_setting.create_time IS '记录创建时间，以时间戳形式存储';
COMMENT ON COLUMN tbl_task_report_setting.update_time IS '记录更新时间，以时间戳形式存储';

-- 创建更新时间触发器
CREATE TRIGGER update_tbl_task_report_setting_timestamp
    BEFORE UPDATE ON tbl_task_report_setting
    FOR EACH ROW
    EXECUTE FUNCTION update_timestamp();
COMMIT;

-- =============================================
-- 任务完成报告表
-- =============================================
BEGIN;
CREATE TABLE tbl_task_report (
    id BIGSERIAL PRIMARY KEY,
    task_id BIGINT NOT NULL,
    assign_id BIGINT NOT NULL,
    report_detail JSONB,
    resource_report_detail JSONB,
    create_time BIGINT DEFAULT EXTRACT(EPOCH FROM CURRENT_TIMESTAMP)::BIGINT,
    update_time BIGINT DEFAULT EXTRACT(EPOCH FROM CURRENT_TIMESTAMP)::BIGINT
);

-- 创建唯一性约束
CREATE UNIQUE INDEX idx_tbl_task_report_task_assign ON tbl_task_report(task_id, assign_id);

-- 表注释：任务完成报告表，记录任务完成情况统计
COMMENT ON TABLE tbl_task_report IS '任务完成报告表，记录任务完成情况统计';
-- 字段注释
COMMENT ON COLUMN tbl_task_report.id IS '自增主键ID';
COMMENT ON COLUMN tbl_task_report.task_id IS '任务ID，关联任务表';
COMMENT ON COLUMN tbl_task_report.assign_id IS '任务布置ID，关联任务布置表';
COMMENT ON COLUMN tbl_task_report.report_detail IS '任务报告详情，JSON格式存储统计信息';
COMMENT ON COLUMN tbl_task_report.resource_report_detail IS '资源维度报告详情，JSON格式存储资源统计信息';
COMMENT ON COLUMN tbl_task_report.create_time IS '记录创建时间';
COMMENT ON COLUMN tbl_task_report.update_time IS '记录更新时间';

-- 创建更新时间触发器
CREATE TRIGGER update_tbl_task_report_timestamp
    BEFORE UPDATE ON tbl_task_report
    FOR EACH ROW
    EXECUTE FUNCTION update_timestamp();
COMMIT;

-- =============================================
-- 学生任务完成进度报告表
-- =============================================
BEGIN;
CREATE TABLE tbl_task_students_report (
    id BIGSERIAL PRIMARY KEY,
    task_id BIGINT NOT NULL,
    assign_id BIGINT NOT NULL,
    student_id BIGINT NOT NULL,
    study_score BIGINT,
    completed_progress REAL,
    accuracy_rate REAL,
    answer_count BIGINT,
    incorrect_count BIGINT,
    cost_time BIGINT,
    task_report JSONB,
    create_time BIGINT DEFAULT EXTRACT(EPOCH FROM CURRENT_TIMESTAMP)::BIGINT,
    update_time BIGINT DEFAULT EXTRACT(EPOCH FROM CURRENT_TIMESTAMP)::BIGINT
);

-- 表注释
COMMENT ON TABLE tbl_task_students_report IS '任务学生报告表';
-- 字段注释
COMMENT ON COLUMN tbl_task_students_report.id IS '主键ID';
COMMENT ON COLUMN tbl_task_students_report.task_id IS '任务ID';
COMMENT ON COLUMN tbl_task_students_report.assign_id IS '布置ID';
COMMENT ON COLUMN tbl_task_students_report.student_id IS '学生ID';
COMMENT ON COLUMN tbl_task_students_report.study_score IS '学习分';
COMMENT ON COLUMN tbl_task_students_report.completed_progress IS '完成进度';
COMMENT ON COLUMN tbl_task_students_report.accuracy_rate IS '正确率';
COMMENT ON COLUMN tbl_task_students_report.answer_count IS '答题数';
COMMENT ON COLUMN tbl_task_students_report.incorrect_count IS '错题数';
COMMENT ON COLUMN tbl_task_students_report.cost_time IS '答题用时';
COMMENT ON COLUMN tbl_task_students_report.task_report IS '分资源统计的报告';
COMMENT ON COLUMN tbl_task_students_report.create_time IS '首次统计时间';
COMMENT ON COLUMN tbl_task_students_report.update_time IS '最后更新时间';

-- 创建索引
CREATE INDEX idx_task_students_report_task_id ON tbl_task_students_report(task_id);
CREATE INDEX idx_task_students_report_assign_id ON tbl_task_students_report(assign_id);
CREATE INDEX idx_task_students_report_student_id ON tbl_task_students_report(student_id);
CREATE INDEX idx_task_students_report_task_assign_student ON tbl_task_students_report(task_id, assign_id, student_id);

-- 创建更新时间触发器
CREATE TRIGGER update_tbl_task_students_report_timestamp
    BEFORE UPDATE ON tbl_task_students_report
    FOR EACH ROW
    EXECUTE FUNCTION update_timestamp();
COMMIT;

-- =============================================
-- 任务完成记录表
-- =============================================
BEGIN;
CREATE TABLE tbl_task_student_details (
    id BIGSERIAL PRIMARY KEY,
    task_id BIGINT NOT NULL,
    assign_id BIGINT NOT NULL DEFAULT 0,
    resource_key VARCHAR(64) NOT NULL,
    question_id VARCHAR(64) NOT NULL,
    student_id BIGINT NOT NULL,
    answer_content TEXT,
    correctness BOOLEAN,
    cost_time BIGINT DEFAULT 0,
    create_time BIGINT DEFAULT EXTRACT(EPOCH FROM CURRENT_TIMESTAMP)::BIGINT,
    update_time BIGINT DEFAULT EXTRACT(EPOCH FROM CURRENT_TIMESTAMP)::BIGINT
);

-- 创建唯一性约束
CREATE UNIQUE INDEX idx_tbl_task_student_details_unique ON tbl_task_student_details(task_id, assign_id, student_id, resource_key, question_id);

-- 表注释
COMMENT ON TABLE tbl_task_student_details IS '学生任务完成详情表';
-- 字段注释
COMMENT ON COLUMN tbl_task_student_details.id IS '自增主键ID';
COMMENT ON COLUMN tbl_task_student_details.task_id IS '任务ID';
COMMENT ON COLUMN tbl_task_student_details.assign_id IS '任务布置ID';
COMMENT ON COLUMN tbl_task_student_details.resource_key IS '资源key resource_id#resource_type';
COMMENT ON COLUMN tbl_task_student_details.question_id IS '题目ID';
COMMENT ON COLUMN tbl_task_student_details.student_id IS '学生ID';
COMMENT ON COLUMN tbl_task_student_details.answer_content IS '作答内容或进度等';
COMMENT ON COLUMN tbl_task_student_details.correctness IS '答案正确性标识';
COMMENT ON COLUMN tbl_task_student_details.cost_time IS '答题用时';
COMMENT ON COLUMN tbl_task_student_details.create_time IS '创建时间';
COMMENT ON COLUMN tbl_task_student_details.update_time IS '更新时间';


-- 创建更新时间触发器
CREATE TRIGGER update_tbl_task_student_details_timestamp
    BEFORE UPDATE ON tbl_task_student_details
    FOR EACH ROW
    EXECUTE FUNCTION update_timestamp();
COMMIT;

-- --------------------------------
-- 教师临时选择表（试题篮、资源篮）
-- --------------------------------
BEGIN;
CREATE TABLE tbl_teacher_temp_selection (
    id BIGSERIAL PRIMARY KEY,
    school_id BIGINT NOT NULL,
    teacher_id BIGINT NOT NULL,
    resource_id VARCHAR(16) NOT NULL,
    resource_type BIGINT NOT NULL,
    create_time BIGINT DEFAULT EXTRACT(EPOCH FROM CURRENT_TIMESTAMP)::BIGINT
);
-- 表注释：教师临时选择表（试题篮、资源篮）
COMMENT ON TABLE tbl_teacher_temp_selection IS '教师临时选择表（试题篮、资源篮）';
-- 字段注释
COMMENT ON COLUMN tbl_teacher_temp_selection.id IS '自增主键ID';
COMMENT ON COLUMN tbl_teacher_temp_selection.school_id IS '学校ID';
COMMENT ON COLUMN tbl_teacher_temp_selection.teacher_id IS '教师ID';
COMMENT ON COLUMN tbl_teacher_temp_selection.resource_id IS '资源ID';
COMMENT ON COLUMN tbl_teacher_temp_selection.resource_type IS '资源类型';
COMMENT ON COLUMN tbl_teacher_temp_selection.create_time IS '记录创建时间';
-- 创建索引
CREATE INDEX idx_tbl_teacher_temp_selection_teacher_id ON tbl_teacher_temp_selection(teacher_id);
-- 创建唯一约束
CREATE UNIQUE INDEX idx_tbl_teacher_temp_selection_unique ON tbl_teacher_temp_selection(teacher_id, resource_id, resource_type);
COMMIT;

