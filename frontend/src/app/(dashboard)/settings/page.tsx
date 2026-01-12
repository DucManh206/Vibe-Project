'use client';

import { useState } from 'react';
import { useTranslations } from 'next-intl';
import { useMutation } from '@tanstack/react-query';
import { useTheme } from 'next-themes';
import { 
  User, 
  Shield, 
  Globe, 
  Palette, 
  Bell,
  Loader2,
  Check
} from 'lucide-react';
import { toast } from 'sonner';

import { Button } from '@/components/ui/button';
import { authApi } from '@/services/api';
import { useAuthStore, useUser } from '@/stores/auth-store';

export default function SettingsPage() {
  const t = useTranslations('settings');
  const tCommon = useTranslations('common');
  const { theme, setTheme } = useTheme();
  const user = useUser();
  const setUser = useAuthStore((state) => state.setUser);

  const [activeTab, setActiveTab] = useState<'profile' | 'security' | 'language' | 'theme' | 'notifications'>('profile');

  // Profile form
  const [email, setEmail] = useState(user?.email || '');

  // Password form
  const [passwords, setPasswords] = useState({
    current: '',
    new: '',
    confirm: '',
  });

  // Update profile mutation
  const updateProfileMutation = useMutation({
    mutationFn: (data: { email?: string }) => authApi.updateMe(data),
    onSuccess: (response) => {
      setUser(response.data);
      toast.success(t('updateSuccess'));
    },
    onError: (error: any) => {
      toast.error(error.response?.data?.message || t('updateError'));
    },
  });

  // Change password mutation
  const changePasswordMutation = useMutation({
    mutationFn: ({ current, newPassword }: { current: string; newPassword: string }) =>
      authApi.changePassword(current, newPassword),
    onSuccess: () => {
      toast.success(t('updateSuccess'));
      setPasswords({ current: '', new: '', confirm: '' });
    },
    onError: (error: any) => {
      toast.error(error.response?.data?.message || t('updateError'));
    },
  });

  const handleUpdateProfile = () => {
    if (email !== user?.email) {
      updateProfileMutation.mutate({ email });
    }
  };

  const handleChangePassword = () => {
    if (passwords.new !== passwords.confirm) {
      toast.error(tCommon('auth.passwordMismatch'));
      return;
    }
    if (passwords.new.length < 8) {
      toast.error(tCommon('auth.passwordTooShort'));
      return;
    }
    changePasswordMutation.mutate({
      current: passwords.current,
      newPassword: passwords.new,
    });
  };

  const tabs = [
    { id: 'profile', label: t('profile'), icon: User },
    { id: 'security', label: t('security'), icon: Shield },
    { id: 'language', label: t('language'), icon: Globe },
    { id: 'theme', label: t('theme'), icon: Palette },
    { id: 'notifications', label: t('notifications'), icon: Bell },
  ] as const;

  return (
    <div className="space-y-6">
      {/* Header */}
      <div>
        <h1 className="text-3xl font-bold">{t('title')}</h1>
        <p className="text-muted-foreground mt-1">
          {t('description')}
        </p>
      </div>

      <div className="grid gap-6 lg:grid-cols-[240px_1fr]">
        {/* Sidebar Tabs */}
        <div className="space-y-1">
          {tabs.map((tab) => (
            <button
              key={tab.id}
              onClick={() => setActiveTab(tab.id)}
              className={`
                w-full flex items-center gap-3 px-4 py-2.5 rounded-lg text-left transition-colors
                ${activeTab === tab.id
                  ? 'bg-primary text-primary-foreground'
                  : 'hover:bg-muted text-muted-foreground hover:text-foreground'
                }
              `}
            >
              <tab.icon className="h-5 w-5" />
              {tab.label}
            </button>
          ))}
        </div>

        {/* Content */}
        <div className="bg-card border rounded-lg p-6">
          {/* Profile Tab */}
          {activeTab === 'profile' && (
            <div className="space-y-6">
              <div>
                <h2 className="text-lg font-semibold">{t('profile')}</h2>
                <p className="text-sm text-muted-foreground">
                  {t('profileDesc')}
                </p>
              </div>

              <div className="space-y-4 max-w-md">
                <div>
                  <label className="block text-sm font-medium mb-1">Email</label>
                  <input
                    type="email"
                    value={email}
                    onChange={(e) => setEmail(e.target.value)}
                    className="w-full px-3 py-2 bg-background border rounded-lg"
                  />
                </div>

                <div>
                  <label className="block text-sm font-medium mb-1">Role</label>
                  <input
                    type="text"
                    value={user?.role || ''}
                    disabled
                    className="w-full px-3 py-2 bg-muted border rounded-lg text-muted-foreground"
                  />
                </div>

                <Button
                  onClick={handleUpdateProfile}
                  disabled={updateProfileMutation.isPending || email === user?.email}
                >
                  {updateProfileMutation.isPending && (
                    <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                  )}
                  {t('updateProfile')}
                </Button>
              </div>
            </div>
          )}

          {/* Security Tab */}
          {activeTab === 'security' && (
            <div className="space-y-6">
              <div>
                <h2 className="text-lg font-semibold">{t('changePassword')}</h2>
                <p className="text-sm text-muted-foreground">
                  {t('securityDesc')}
                </p>
              </div>

              <div className="space-y-4 max-w-md">
                <div>
                  <label className="block text-sm font-medium mb-1">
                    {t('currentPassword')}
                  </label>
                  <input
                    type="password"
                    value={passwords.current}
                    onChange={(e) => setPasswords({ ...passwords, current: e.target.value })}
                    className="w-full px-3 py-2 bg-background border rounded-lg"
                  />
                </div>

                <div>
                  <label className="block text-sm font-medium mb-1">
                    {t('newPassword')}
                  </label>
                  <input
                    type="password"
                    value={passwords.new}
                    onChange={(e) => setPasswords({ ...passwords, new: e.target.value })}
                    className="w-full px-3 py-2 bg-background border rounded-lg"
                  />
                </div>

                <div>
                  <label className="block text-sm font-medium mb-1">
                    {t('confirmNewPassword')}
                  </label>
                  <input
                    type="password"
                    value={passwords.confirm}
                    onChange={(e) => setPasswords({ ...passwords, confirm: e.target.value })}
                    className="w-full px-3 py-2 bg-background border rounded-lg"
                  />
                </div>

                <Button
                  onClick={handleChangePassword}
                  disabled={changePasswordMutation.isPending || !passwords.current || !passwords.new || !passwords.confirm}
                >
                  {changePasswordMutation.isPending && (
                    <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                  )}
                  {t('changePassword')}
                </Button>
              </div>
            </div>
          )}

          {/* Language Tab */}
          {activeTab === 'language' && (
            <div className="space-y-6">
              <div>
                <h2 className="text-lg font-semibold">{t('language')}</h2>
                <p className="text-sm text-muted-foreground">
                  {t('languageDesc')}
                </p>
              </div>

              <div className="space-y-2 max-w-md">
                <button
                  onClick={() => {
                    document.cookie = `NEXT_LOCALE=vi; path=/; max-age=31536000`;
                    window.location.reload();
                  }}
                  className={`w-full flex items-center justify-between p-4 border rounded-lg hover:bg-muted transition-colors ${
                    tCommon('appName') !== 'Captcha Platform' ? 'border-primary bg-primary/5' : ''
                  }`}
                >
                  <div className="flex items-center gap-3">
                    <span className="text-2xl">🇻🇳</span>
                    <span>Tiếng Việt</span>
                  </div>
                  {tCommon('appName') !== 'Captcha Platform' && <Check className="h-5 w-5 text-primary" />}
                </button>

                <button
                  onClick={() => {
                    document.cookie = `NEXT_LOCALE=en; path=/; max-age=31536000`;
                    window.location.reload();
                  }}
                  className={`w-full flex items-center justify-between p-4 border rounded-lg hover:bg-muted transition-colors ${
                    tCommon('appName') === 'Captcha Platform' ? 'border-primary bg-primary/5' : ''
                  }`}
                >
                  <div className="flex items-center gap-3">
                    <span className="text-2xl">🇺🇸</span>
                    <span>English</span>
                  </div>
                  {tCommon('appName') === 'Captcha Platform' && <Check className="h-5 w-5 text-primary" />}
                </button>
              </div>
            </div>
          )}

          {/* Theme Tab */}
          {activeTab === 'theme' && (
            <div className="space-y-6">
              <div>
                <h2 className="text-lg font-semibold">{t('theme')}</h2>
                <p className="text-sm text-muted-foreground">
                  {t('themeDesc')}
                </p>
              </div>

              <div className="grid gap-4 max-w-md">
                <button
                  onClick={() => setTheme('light')}
                  className={`flex items-center justify-between p-4 border rounded-lg transition-colors ${
                    theme === 'light' ? 'border-primary bg-primary/5' : 'hover:bg-muted'
                  }`}
                >
                  <div className="flex items-center gap-3">
                    <div className="w-10 h-10 rounded-lg bg-white border flex items-center justify-center">
                      ☀️
                    </div>
                    <span>{t('lightTheme')}</span>
                  </div>
                  {theme === 'light' && <Check className="h-5 w-5 text-primary" />}
                </button>

                <button
                  onClick={() => setTheme('dark')}
                  className={`flex items-center justify-between p-4 border rounded-lg transition-colors ${
                    theme === 'dark' ? 'border-primary bg-primary/5' : 'hover:bg-muted'
                  }`}
                >
                  <div className="flex items-center gap-3">
                    <div className="w-10 h-10 rounded-lg bg-gray-900 border flex items-center justify-center">
                      🌙
                    </div>
                    <span>{t('darkTheme')}</span>
                  </div>
                  {theme === 'dark' && <Check className="h-5 w-5 text-primary" />}
                </button>

                <button
                  onClick={() => setTheme('system')}
                  className={`flex items-center justify-between p-4 border rounded-lg transition-colors ${
                    theme === 'system' ? 'border-primary bg-primary/5' : 'hover:bg-muted'
                  }`}
                >
                  <div className="flex items-center gap-3">
                    <div className="w-10 h-10 rounded-lg bg-gradient-to-br from-white to-gray-900 border flex items-center justify-center">
                      💻
                    </div>
                    <span>{t('systemTheme')}</span>
                  </div>
                  {theme === 'system' && <Check className="h-5 w-5 text-primary" />}
                </button>
              </div>
            </div>
          )}

          {/* Notifications Tab */}
          {activeTab === 'notifications' && (
            <div className="space-y-6">
              <div>
                <h2 className="text-lg font-semibold">{t('notifications')}</h2>
                <p className="text-sm text-muted-foreground">
                  {t('notificationsDesc')}
                </p>
              </div>

              <div className="space-y-4 max-w-md">
                <div className="flex items-center justify-between p-4 border rounded-lg">
                  <div>
                    <p className="font-medium">{t('notifEmail')}</p>
                    <p className="text-sm text-muted-foreground">
                      {t('notifEmailDesc')}
                    </p>
                  </div>
                  <input type="checkbox" defaultChecked className="h-5 w-5" />
                </div>

                <div className="flex items-center justify-between p-4 border rounded-lg">
                  <div>
                    <p className="font-medium">{t('notifTraining')}</p>
                    <p className="text-sm text-muted-foreground">
                      {t('notifTrainingDesc')}
                    </p>
                  </div>
                  <input type="checkbox" defaultChecked className="h-5 w-5" />
                </div>

                <div className="flex items-center justify-between p-4 border rounded-lg">
                  <div>
                    <p className="font-medium">{t('notifSecurity')}</p>
                    <p className="text-sm text-muted-foreground">
                      {t('notifSecurityDesc')}
                    </p>
                  </div>
                  <input type="checkbox" defaultChecked className="h-5 w-5" />
                </div>
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}