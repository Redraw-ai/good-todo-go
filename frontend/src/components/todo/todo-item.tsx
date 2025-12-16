'use client';

import { useState } from 'react';
import { Checkbox } from '@/components/ui/checkbox';
import { Button } from '@/components/ui/button';
import { Card, CardContent } from '@/components/ui/card';
import { Trash2, Globe, Lock } from 'lucide-react';
import { TodoResponse } from '@/api/public/model/components-schemas-todo';
import { useUpdateTodo, useDeleteTodo, getGetTodosQueryKey, getGetPublicTodosQueryKey } from '@/api/public/todo/todo';
import { useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { format } from 'date-fns';
import { ja } from 'date-fns/locale';

interface TodoItemProps {
  todo: TodoResponse;
  isOwner?: boolean;
}

export function TodoItem({ todo, isOwner = false }: TodoItemProps) {
  const queryClient = useQueryClient();
  const [isDeleting, setIsDeleting] = useState(false);

  const updateTodo = useUpdateTodo();
  const deleteTodo = useDeleteTodo();

  const handleToggleComplete = async () => {
    if (!todo.id || !isOwner) return;

    try {
      await updateTodo.mutateAsync({
        todoId: todo.id,
        data: { completed: !todo.completed },
      });
      queryClient.invalidateQueries({ queryKey: getGetTodosQueryKey() });
      queryClient.invalidateQueries({ queryKey: getGetPublicTodosQueryKey() });
    } catch (error) {
      toast.error('更新に失敗しました');
      console.error(error);
    }
  };

  const handleTogglePublic = async () => {
    if (!todo.id || !isOwner) return;

    try {
      await updateTodo.mutateAsync({
        todoId: todo.id,
        data: { is_public: !todo.is_public },
      });
      queryClient.invalidateQueries({ queryKey: getGetTodosQueryKey() });
      queryClient.invalidateQueries({ queryKey: getGetPublicTodosQueryKey() });
      toast.success(todo.is_public ? '非公開にしました' : '公開しました');
    } catch (error) {
      toast.error('更新に失敗しました');
      console.error(error);
    }
  };

  const handleDelete = async () => {
    if (!todo.id || !isOwner) return;

    setIsDeleting(true);
    try {
      await deleteTodo.mutateAsync({ todoId: todo.id });
      queryClient.invalidateQueries({ queryKey: getGetTodosQueryKey() });
      queryClient.invalidateQueries({ queryKey: getGetPublicTodosQueryKey() });
      toast.success('削除しました');
    } catch (error) {
      toast.error('削除に失敗しました');
      console.error(error);
    } finally {
      setIsDeleting(false);
    }
  };

  return (
    <Card className={`${todo.completed ? 'opacity-60' : ''}`}>
      <CardContent className="flex items-center gap-4 p-4">
        {isOwner ? (
          <Checkbox
            checked={todo.completed ?? false}
            onCheckedChange={handleToggleComplete}
            disabled={updateTodo.isPending}
          />
        ) : (
          <div className="w-4 h-4 rounded border border-gray-300 flex items-center justify-center">
            {todo.completed && <div className="w-2 h-2 bg-gray-400 rounded-sm" />}
          </div>
        )}

        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2">
            <h3 className={`font-medium truncate ${todo.completed ? 'line-through text-gray-500' : ''}`}>
              {todo.title}
            </h3>
            <span title={todo.is_public ? '公開' : '非公開'}>
              {todo.is_public ? (
                <Globe className="h-4 w-4 text-blue-500 shrink-0" />
              ) : (
                <Lock className="h-4 w-4 text-gray-400 shrink-0" />
              )}
            </span>
          </div>
          {todo.description && (
            <p className="text-sm text-gray-500 truncate">{todo.description}</p>
          )}
          {todo.due_date && (
            <p className="text-xs text-gray-400 mt-1">
              期限: {format(new Date(todo.due_date), 'yyyy年M月d日', { locale: ja })}
            </p>
          )}
        </div>

        {isOwner && (
          <div className="flex items-center gap-2">
            <Button
              variant="ghost"
              size="icon"
              onClick={handleTogglePublic}
              disabled={updateTodo.isPending}
              title={todo.is_public ? '非公開にする' : '公開する'}
            >
              {todo.is_public ? (
                <Lock className="h-4 w-4" />
              ) : (
                <Globe className="h-4 w-4" />
              )}
            </Button>
            <Button
              variant="ghost"
              size="icon"
              onClick={handleDelete}
              disabled={isDeleting}
              className="text-red-500 hover:text-red-700"
              title="削除"
            >
              <Trash2 className="h-4 w-4" />
            </Button>
          </div>
        )}
      </CardContent>
    </Card>
  );
}
