package types

type Story struct {
	Chat Chat `json:"chat"`
	ID   int  `json:"id"`
}

type ChecklistTask struct {
	ID              int             `json:"id"`
	Text            string          `json:"text"`
	TextEntities    []MessageEntity `json:"text_entities,omitempty"`
	CompletedByUser *User           `json:"completed_by_user,omitempty"`
	CompletedByChat *Chat           `json:"completed_by_chat,omitempty"`
	CompletionDate  int64           `json:"completion_date,omitempty"`
}

type Checklist struct {
	Title                    string          `json:"title"`
	TitleEntities            []MessageEntity `json:"title_entities,omitempty"`
	Tasks                    []ChecklistTask `json:"tasks"`
	OthersCanAddTasks        bool            `json:"others_can_add_tasks,omitempty"`
	OthersCanMarkTasksAsDone bool            `json:"others_can_mark_tasks_as_done,omitempty"`
}
