'use client';

import { TodoItem } from './todo-item';
import { TodoResponse } from '@/api/public/model/components-schemas-todo';

interface TodoListProps {
  todos: TodoResponse[];
  currentUserId?: string;
  emptyMessage?: string;
}

export function TodoList({ todos, currentUserId, emptyMessage = 'Todoがありません' }: TodoListProps) {
  if (todos.length === 0) {
    return (
      <div className="text-center py-8 text-gray-500">
        {emptyMessage}
      </div>
    );
  }

  return (
    <div className="space-y-3">
      {todos.map((todo) => (
        <TodoItem
          key={todo.id}
          todo={todo}
          isOwner={currentUserId === todo.user_id}
        />
      ))}
    </div>
  );
}
