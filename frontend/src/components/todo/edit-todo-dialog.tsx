'use client';

import { useEffect } from 'react';
import { useForm } from 'react-hook-form';
import { useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Checkbox } from '@/components/ui/checkbox';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { useUpdateTodo, getGetTodosQueryKey, getGetPublicTodosQueryKey } from '@/api/public/todo/todo';
import { TodoResponse } from '@/api/public/model/components-schemas-todo';
import { format } from 'date-fns';

interface EditTodoFormData {
  title: string;
  description?: string;
  is_public: boolean;
  due_date?: string;
}

interface EditTodoDialogProps {
  todo: TodoResponse;
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export function EditTodoDialog({ todo, open, onOpenChange }: EditTodoDialogProps) {
  const queryClient = useQueryClient();

  const { register, handleSubmit, reset, formState: { errors }, watch, setValue } = useForm<EditTodoFormData>({
    defaultValues: {
      title: todo.title ?? '',
      description: todo.description ?? '',
      is_public: todo.is_public ?? false,
      due_date: todo.due_date ? format(new Date(todo.due_date), 'yyyy-MM-dd') : '',
    },
  });

  const updateTodo = useUpdateTodo();
  const isPublic = watch('is_public');

  useEffect(() => {
    if (open) {
      reset({
        title: todo.title ?? '',
        description: todo.description ?? '',
        is_public: todo.is_public ?? false,
        due_date: todo.due_date ? format(new Date(todo.due_date), 'yyyy-MM-dd') : '',
      });
    }
  }, [open, todo, reset]);

  const onSubmit = async (data: EditTodoFormData) => {
    if (!todo.id) return;

    try {
      await updateTodo.mutateAsync({
        todoId: todo.id,
        data: {
          title: data.title,
          description: data.description || undefined,
          is_public: data.is_public,
          due_date: data.due_date ? new Date(data.due_date).toISOString() : undefined,
        },
      });
      queryClient.invalidateQueries({ queryKey: getGetTodosQueryKey() });
      queryClient.invalidateQueries({ queryKey: getGetPublicTodosQueryKey() });
      toast.success('Todoを更新しました');
      onOpenChange(false);
    } catch (error) {
      toast.error('Todoの更新に失敗しました');
      console.error(error);
    }
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[425px]">
        <form onSubmit={handleSubmit(onSubmit)}>
          <DialogHeader>
            <DialogTitle>Todoを編集</DialogTitle>
            <DialogDescription>
              Todoの内容を編集してください。
            </DialogDescription>
          </DialogHeader>
          <div className="grid gap-4 py-4">
            <div className="space-y-2">
              <Label htmlFor="edit-title">タイトル</Label>
              <Input
                id="edit-title"
                placeholder="タイトルを入力"
                {...register('title', { required: 'タイトルは必須です' })}
              />
              {errors.title && (
                <p className="text-sm text-red-500">{errors.title.message}</p>
              )}
            </div>
            <div className="space-y-2">
              <Label htmlFor="edit-description">説明（任意）</Label>
              <Input
                id="edit-description"
                placeholder="説明を入力"
                {...register('description')}
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="edit-due_date">期限（任意）</Label>
              <Input
                id="edit-due_date"
                type="date"
                {...register('due_date')}
              />
            </div>
            <div className="flex items-center space-x-2">
              <Checkbox
                id="edit-is_public"
                checked={isPublic}
                onCheckedChange={(checked) => setValue('is_public', checked === true)}
              />
              <Label htmlFor="edit-is_public" className="text-sm font-normal">
                チームに公開する
              </Label>
            </div>
          </div>
          <DialogFooter>
            <Button type="button" variant="outline" onClick={() => onOpenChange(false)}>
              キャンセル
            </Button>
            <Button type="submit" disabled={updateTodo.isPending}>
              {updateTodo.isPending ? '更新中...' : '更新'}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
