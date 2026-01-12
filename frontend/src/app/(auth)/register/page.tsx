'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import Link from 'next/link';
import { useTranslations } from 'next-intl';
import { useMutation } from '@tanstack/react-query';
import { Shield, Loader2, Mail, Lock, Eye, EyeOff, Check } from 'lucide-react';
import { toast } from 'sonner';

import { Button } from '@/components/ui/button';
import { authApi } from '@/services/api';

export default function RegisterPage() {
  const t = useTranslations('auth');
  const router = useRouter();

  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [showPassword, setShowPassword] = useState(false);
  const [agreed, setAgreed] = useState(false);

  const registerMutation = useMutation({
    mutationFn: authApi.register,
    onSuccess: () => {
      toast.success(t('registerSuccess'));
      router.push('/login');
    },
    onError: (error: any) => {
      toast.error(error.response?.data?.message || t('registerError'));
    },
  });

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();

    if (!email) {
      toast.error(t('emailRequired'));
      return;
    }
    if (!password) {
      toast.error(t('passwordRequired'));
      return;
    }
    if (password.length < 8) {
      toast.error(t('passwordTooShort'));
      return;
    }
    if (password !== confirmPassword) {
      toast.error(t('passwordMismatch'));
      return;
    }
    if (!agreed) {
      toast.error('Vui lòng đồng ý với điều khoản sử dụng');
      return;
    }

    registerMutation.mutate({ email, password });
  };

  // Password strength indicator
  const getPasswordStrength = () => {
    if (!password) return { strength: 0, text: '', color: '' };
    
    let strength = 0;
    if (password.length >= 8) strength++;
    if (/[A-Z]/.test(password)) strength++;
    if (/[a-z]/.test(password)) strength++;
    if (/[0-9]/.test(password)) strength++;
    if (/[^A-Za-z0-9]/.test(password)) strength++;

    const levels = [
      { text: 'Rất yếu', color: 'bg-red-500' },
      { text: 'Yếu', color: 'bg-orange-500' },
      { text: 'Trung bình', color: 'bg-yellow-500' },
      { text: 'Mạnh', color: 'bg-lime-500' },
      { text: 'Rất mạnh', color: 'bg-green-500' },
    ];

    return { strength, ...levels[Math.min(strength - 1, 4)] };
  };

  const passwordStrength = getPasswordStrength();

  return (
    <div className="min-h-screen flex items-center justify-center bg-gradient-to-b from-background to-muted px-4 py-12">
      <div className="w-full max-w-md">
        {/* Logo */}
        <div className="text-center mb-8">
          <Link href="/" className="inline-flex items-center gap-2">
            <Shield className="h-10 w-10 text-primary" />
            <span className="text-2xl font-bold">Captcha Platform</span>
          </Link>
        </div>

        {/* Register Card */}
        <div className="bg-card border rounded-xl p-8 shadow-lg">
          <h1 className="text-2xl font-bold text-center mb-2">{t('register')}</h1>
          <p className="text-muted-foreground text-center mb-6">
            Tạo tài khoản mới để bắt đầu
          </p>

          <form onSubmit={handleSubmit} className="space-y-4">
            {/* Email */}
            <div>
              <label className="block text-sm font-medium mb-1.5">{t('email')}</label>
              <div className="relative">
                <Mail className="absolute left-3 top-1/2 -translate-y-1/2 h-5 w-5 text-muted-foreground" />
                <input
                  type="email"
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                  className="w-full pl-10 pr-4 py-2.5 bg-background border rounded-lg focus:ring-2 focus:ring-primary focus:border-primary"
                  placeholder="email@example.com"
                  autoComplete="email"
                />
              </div>
            </div>

            {/* Password */}
            <div>
              <label className="block text-sm font-medium mb-1.5">{t('password')}</label>
              <div className="relative">
                <Lock className="absolute left-3 top-1/2 -translate-y-1/2 h-5 w-5 text-muted-foreground" />
                <input
                  type={showPassword ? 'text' : 'password'}
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  className="w-full pl-10 pr-12 py-2.5 bg-background border rounded-lg focus:ring-2 focus:ring-primary focus:border-primary"
                  placeholder="••••••••"
                  autoComplete="new-password"
                />
                <button
                  type="button"
                  onClick={() => setShowPassword(!showPassword)}
                  className="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground"
                >
                  {showPassword ? <EyeOff className="h-5 w-5" /> : <Eye className="h-5 w-5" />}
                </button>
              </div>
              
              {/* Password strength */}
              {password && (
                <div className="mt-2">
                  <div className="flex gap-1 mb-1">
                    {[1, 2, 3, 4, 5].map((i) => (
                      <div
                        key={i}
                        className={`h-1 flex-1 rounded-full ${
                          i <= passwordStrength.strength ? passwordStrength.color : 'bg-muted'
                        }`}
                      />
                    ))}
                  </div>
                  <p className="text-xs text-muted-foreground">{passwordStrength.text}</p>
                </div>
              )}
            </div>

            {/* Confirm Password */}
            <div>
              <label className="block text-sm font-medium mb-1.5">{t('confirmPassword')}</label>
              <div className="relative">
                <Lock className="absolute left-3 top-1/2 -translate-y-1/2 h-5 w-5 text-muted-foreground" />
                <input
                  type={showPassword ? 'text' : 'password'}
                  value={confirmPassword}
                  onChange={(e) => setConfirmPassword(e.target.value)}
                  className="w-full pl-10 pr-12 py-2.5 bg-background border rounded-lg focus:ring-2 focus:ring-primary focus:border-primary"
                  placeholder="••••••••"
                  autoComplete="new-password"
                />
                {confirmPassword && confirmPassword === password && (
                  <Check className="absolute right-3 top-1/2 -translate-y-1/2 h-5 w-5 text-green-500" />
                )}
              </div>
            </div>

            {/* Terms */}
            <div>
              <label className="flex items-start gap-2 text-sm">
                <input
                  type="checkbox"
                  checked={agreed}
                  onChange={(e) => setAgreed(e.target.checked)}
                  className="rounded border-gray-300 mt-0.5"
                />
                <span className="text-muted-foreground">
                  Tôi đồng ý với{' '}
                  <Link href="/terms" className="text-primary hover:underline">
                    Điều khoản sử dụng
                  </Link>{' '}
                  và{' '}
                  <Link href="/privacy" className="text-primary hover:underline">
                    Chính sách bảo mật
                  </Link>
                </span>
              </label>
            </div>

            {/* Submit Button */}
            <Button
              type="submit"
              className="w-full"
              size="lg"
              disabled={registerMutation.isPending}
            >
              {registerMutation.isPending ? (
                <>
                  <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                  Đang đăng ký...
                </>
              ) : (
                t('register')
              )}
            </Button>
          </form>

          {/* Login Link */}
          <p className="text-center mt-6 text-sm text-muted-foreground">
            {t('hasAccount')}{' '}
            <Link href="/login" className="text-primary hover:underline font-medium">
              {t('login')}
            </Link>
          </p>
        </div>
      </div>
    </div>
  );
}