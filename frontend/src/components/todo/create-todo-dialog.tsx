'use client';

import { useState } from 'react';
import { useForm } from 'react-hook-form';
import { useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { Plus } from 'lucide-react';
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
  DialogTrigger,
} from '@/components/ui/dialog';
import { useCreateTodo, getGetTodosQueryKey, getGetPublicTodosQueryKey } from '@/api/public/todo/todo';

interface CreateTodoFormData {
  title: string;
  description?: string;
  is_public: boolean;
  due_date?: string;
}

export function CreateTodoDialog() {
  const [open, setOpen] = useState(false);
  const queryClient = useQueryClient();

  const { register, handleSubmit, reset, formState: { errors }, watch, setValue } = useForm<CreateTodoFormData>({
    defaultValues: {
      is_public: false,
    },
  });

  const createTodo = useCreateTodo();
  const isPublic = watch('is_public');

  const onSubmit = async (data: CreateTodoFormData) => {
    try {
      await createTodo.mutateAsync({
        data: {
          title: data.title,
          description: data.description,
          is_public: data.is_public,
          due_date: data.due_date ? new Date(data.due_date).toISOString() : undefined,
        },
      });
      queryClient.invalidateQueries({ queryKey: getGetTodosQueryKey() });
      queryClient.invalidateQueries({ queryKey: getGetPublicTodosQueryKey() });
      toast.success('Todoを作成しました');
      reset();
      setOpen(false);
    } catch (error) {
      toast.error('Todoの作成に失敗しました');
      console.error(error);
    }
  };

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        <Button>
          <Plus className="h-4 w-4 mr-2" />
          新規作成
        </Button>
      </DialogTrigger>
      <DialogContent className="sm:max-w-[425px]">
        <form onSubmit={handleSubmit(onSubmit)}>
          <DialogHeader>
            <DialogTitle>新しいTodoを作成</DialogTitle>
            <DialogDescription>
              Todoの内容を入力してください。
            </DialogDescription>
          </DialogHeader>
          <div className="grid gap-4 py-4">
            <div className="space-y-2">
              <Label htmlFor="title">タイトル</Label>
              <Input
                id="title"
                placeholder="タイトルを入力"
                {...register('title', { required: 'タイトルは必須です' })}
              />
              {errors.title && (
                <p className="text-sm text-red-500">{errors.title.message}</p>
              )}
            </div>
            <div className="space-y-2">
              <Label htmlFor="description">説明（任意）</Label>
              <Input
                id="description"
                placeholder="説明を入力"
                {...register('description')}
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="due_date">期限（任意）</Label>
              <Input
                id="due_date"
                type="date"
                {...register('due_date')}
              />
            </div>
            <div className="flex items-center space-x-2">
              <Checkbox
                id="is_public"
                checked={isPublic}
                onCheckedChange={(checked) => setValue('is_public', checked === true)}
              />
              <Label htmlFor="is_public" className="text-sm font-normal">
                チームに公開する
              </Label>
            </div>
          </div>
          <DialogFooter>
            <Button type="button" variant="outline" onClick={() => setOpen(false)}>
              キャンセル
            </Button>
            <Button type="submit" disabled={createTodo.isPending}>
              {createTodo.isPending ? '作成中...' : '作成'}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
