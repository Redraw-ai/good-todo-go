package output

import (
	"good-todo-go/internal/domain/model"
)

type TodoOutput struct {
	ID          string
	Title       string
	Description string
	Completed   bool
	DueDate     *string
	CreatedAt   string
	UpdatedAt   string
}

type TodoListOutput struct {
	Todos []*TodoOutput
	Total int
}

func NewTodoOutput(todo *model.Todo) *TodoOutput {
	var dueDate *string
	if todo.DueDate != nil {
		formatted := todo.DueDate.Format("2006-01-02T15:04:05Z07:00")
		dueDate = &formatted
	}

	return &TodoOutput{
		ID:          todo.ID,
		Title:       todo.Title,
		Description: todo.Description,
		Completed:   todo.Completed,
		DueDate:     dueDate,
		CreatedAt:   todo.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   todo.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func NewTodoListOutput(todos []*model.Todo, total int) *TodoListOutput {
	outputs := make([]*TodoOutput, len(todos))
	for i, t := range todos {
		outputs[i] = NewTodoOutput(t)
	}

	return &TodoListOutput{
		Todos: outputs,
		Total: total,
	}
}
