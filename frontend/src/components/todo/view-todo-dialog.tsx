'use client';

import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { TodoResponse } from '@/api/public/model/components-schemas-todo';
import { format } from 'date-fns';
import { ja } from 'date-fns/locale';
import { Globe, Calendar, User, CheckCircle, Circle } from 'lucide-react';

interface ViewTodoDialogProps {
  todo: TodoResponse;
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export function ViewTodoDialog({ todo, open, onOpenChange }: ViewTodoDialogProps) {
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[425px]">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            {todo.completed ? (
              <CheckCircle className="h-5 w-5 text-green-500" />
            ) : (
              <Circle className="h-5 w-5 text-gray-400" />
            )}
            <span className={todo.completed ? 'line-through text-gray-500' : ''}>
              {todo.title}
            </span>
          </DialogTitle>
          <DialogDescription>
            チームメンバーのTodo
          </DialogDescription>
        </DialogHeader>
        <div className="space-y-4 py-4">
          {todo.description && (
            <div className="space-y-1">
              <p className="text-sm font-medium text-gray-500">説明</p>
              <p className="text-sm">{todo.description}</p>
            </div>
          )}

          {todo.created_by && (
            <div className="flex items-center gap-2 text-sm">
              <User className="h-4 w-4 text-gray-400" />
              <span className="text-gray-500">作成者:</span>
              <span>{todo.created_by.name}</span>
            </div>
          )}

          {todo.due_date && (
            <div className="flex items-center gap-2 text-sm">
              <Calendar className="h-4 w-4 text-gray-400" />
              <span className="text-gray-500">期限:</span>
              <span>{format(new Date(todo.due_date), 'yyyy年M月d日', { locale: ja })}</span>
            </div>
          )}

          <div className="flex items-center gap-2 text-sm">
            <Globe className="h-4 w-4 text-blue-500" />
            <span className="text-gray-500">公開設定:</span>
            <span>チームに公開中</span>
          </div>

          {todo.created_at && (
            <div className="text-xs text-gray-400">
              作成日: {format(new Date(todo.created_at), 'yyyy年M月d日 HH:mm', { locale: ja })}
            </div>
          )}
        </div>
      </DialogContent>
    </Dialog>
  );
}
