'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import Link from 'next/link';
import { useForm } from 'react-hook-form';
import { toast } from 'sonner';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from '@/components/ui/card';
import { useAuth } from '@/contexts/auth-context';
import { useRegister } from '@/api/public/auth/auth';

interface RegisterFormData {
  name: string;
  email: string;
  password: string;
  confirmPassword: string;
  tenant_slug: string;
}

export default function RegisterPage() {
  const router = useRouter();
  const { login } = useAuth();
  const [isLoading, setIsLoading] = useState(false);

  const { register: registerField, handleSubmit, formState: { errors }, watch } = useForm<RegisterFormData>();

  const registerMutation = useRegister();

  const password = watch('password');

  const onSubmit = async (data: RegisterFormData) => {
    setIsLoading(true);
    try {
      const response = await registerMutation.mutateAsync({
        data: {
          name: data.name,
          email: data.email,
          password: data.password,
          tenant_slug: data.tenant_slug,
        },
      });
      if (response.access_token && response.refresh_token) {
        login(response.access_token, response.refresh_token, data.tenant_slug);
        toast.success('登録が完了しました');
        router.push('/todos');
      }
    } catch (error) {
      toast.error('登録に失敗しました。もう一度お試しください。');
      console.error('Registration error:', error);
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50 px-4">
      <Card className="w-full max-w-md">
        <CardHeader className="space-y-1">
          <CardTitle className="text-2xl font-bold text-center">新規登録</CardTitle>
          <CardDescription className="text-center">
            アカウントを作成して始めましょう
          </CardDescription>
        </CardHeader>
        <form onSubmit={handleSubmit(onSubmit)}>
          <CardContent className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="tenant_slug">組織ID</Label>
              <Input
                id="tenant_slug"
                type="text"
                placeholder="your-organization"
                {...registerField('tenant_slug', {
                  required: '組織IDは必須です',
                  pattern: {
                    value: /^[a-z0-9-]+$/,
                    message: '小文字英数字とハイフンのみ使用できます',
                  },
                })}
              />
              {errors.tenant_slug && (
                <p className="text-sm text-red-500">{errors.tenant_slug.message}</p>
              )}
            </div>
            <div className="space-y-2">
              <Label htmlFor="name">名前</Label>
              <Input
                id="name"
                type="text"
                placeholder="山田 太郎"
                {...registerField('name', {
                  required: '名前は必須です',
                  minLength: {
                    value: 2,
                    message: '名前は2文字以上で入力してください',
                  },
                })}
              />
              {errors.name && (
                <p className="text-sm text-red-500">{errors.name.message}</p>
              )}
            </div>
            <div className="space-y-2">
              <Label htmlFor="email">メールアドレス</Label>
              <Input
                id="email"
                type="email"
                placeholder="name@example.com"
                {...registerField('email', {
                  required: 'メールアドレスは必須です',
                  pattern: {
                    value: /^[A-Z0-9._%+-]+@[A-Z0-9.-]+\.[A-Z]{2,}$/i,
                    message: '有効なメールアドレスを入力してください',
                  },
                })}
              />
              {errors.email && (
                <p className="text-sm text-red-500">{errors.email.message}</p>
              )}
            </div>
            <div className="space-y-2">
              <Label htmlFor="password">パスワード</Label>
              <Input
                id="password"
                type="password"
                {...registerField('password', {
                  required: 'パスワードは必須です',
                  minLength: {
                    value: 8,
                    message: 'パスワードは8文字以上で入力してください',
                  },
                })}
              />
              {errors.password && (
                <p className="text-sm text-red-500">{errors.password.message}</p>
              )}
            </div>
            <div className="space-y-2">
              <Label htmlFor="confirmPassword">パスワード（確認）</Label>
              <Input
                id="confirmPassword"
                type="password"
                {...registerField('confirmPassword', {
                  required: 'パスワードを再入力してください',
                  validate: (value) =>
                    value === password || 'パスワードが一致しません',
                })}
              />
              {errors.confirmPassword && (
                <p className="text-sm text-red-500">{errors.confirmPassword.message}</p>
              )}
            </div>
          </CardContent>
          <CardFooter className="flex flex-col space-y-4">
            <Button type="submit" className="w-full" disabled={isLoading}>
              {isLoading ? '登録中...' : '登録'}
            </Button>
            <p className="text-sm text-center text-gray-600">
              すでにアカウントをお持ちの方は{' '}
              <Link href="/login" className="text-blue-600 hover:underline">
                ログイン
              </Link>
            </p>
          </CardFooter>
        </form>
      </Card>
    </div>
  );
}
