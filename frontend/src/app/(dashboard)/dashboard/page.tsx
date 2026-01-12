'use client';

import { useTranslations } from 'next-intl';
import { useQuery } from '@tanstack/react-query';
import { 
  Activity, 
  CheckCircle, 
  XCircle, 
  Clock, 
  Cpu,
  TrendingUp,
  Image,
  Key
} from 'lucide-react';

import { captchaApi, apiKeysApi } from '@/services/api';
import { useUser } from '@/stores/auth-store';

export default function DashboardPage() {
  const t = useTranslations('dashboard');
  const user = useUser();

  const { data: stats, isLoading: statsLoading } = useQuery({
    queryKey: ['stats'],
    queryFn: () => captchaApi.getStats(),
  });

  const { data: apiKeys } = useQuery({
    queryKey: ['apiKeys'],
    queryFn: () => apiKeysApi.list(),
  });

  const statsData = stats?.data;

  return (
    <div className="space-y-8">
      {/* Header */}
      <div>
        <h1 className="text-3xl font-bold">{t('title')}</h1>
        <p className="text-muted-foreground mt-1">
          {t('welcome')}, {user?.email}
        </p>
      </div>

      {/* Stats Grid */}
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        {/* Total Requests */}
        <div className="bg-card border rounded-lg p-6">
          <div className="flex items-center gap-4">
            <div className="p-3 rounded-full bg-blue-500/10">
              <Activity className="h-6 w-6 text-blue-500" />
            </div>
            <div>
              <p className="text-sm text-muted-foreground">{t('totalRequests')}</p>
              <p className="text-2xl font-bold">
                {statsLoading ? '...' : statsData?.total_requests?.toLocaleString() || '0'}
              </p>
            </div>
          </div>
        </div>

        {/* Success Rate */}
        <div className="bg-card border rounded-lg p-6">
          <div className="flex items-center gap-4">
            <div className="p-3 rounded-full bg-green-500/10">
              <CheckCircle className="h-6 w-6 text-green-500" />
            </div>
            <div>
              <p className="text-sm text-muted-foreground">{t('successRate')}</p>
              <p className="text-2xl font-bold">
                {statsLoading ? '...' : `${((statsData?.accuracy_rate || 0) * 100).toFixed(1)}%`}
              </p>
            </div>
          </div>
        </div>

        {/* Avg Processing Time */}
        <div className="bg-card border rounded-lg p-6">
          <div className="flex items-center gap-4">
            <div className="p-3 rounded-full bg-yellow-500/10">
              <Clock className="h-6 w-6 text-yellow-500" />
            </div>
            <div>
              <p className="text-sm text-muted-foreground">{t('avgProcessingTime')}</p>
              <p className="text-2xl font-bold">
                {statsLoading ? '...' : `${statsData?.average_processing_time_ms?.toFixed(0) || '0'}ms`}
              </p>
            </div>
          </div>
        </div>

        {/* Active Models */}
        <div className="bg-card border rounded-lg p-6">
          <div className="flex items-center gap-4">
            <div className="p-3 rounded-full bg-purple-500/10">
              <Cpu className="h-6 w-6 text-purple-500" />
            </div>
            <div>
              <p className="text-sm text-muted-foreground">{t('activeModels')}</p>
              <p className="text-2xl font-bold">
                {statsLoading ? '...' : statsData?.active_models_count || '0'}
              </p>
            </div>
          </div>
        </div>
      </div>

      {/* Quick Actions & Recent Activity */}
      <div className="grid gap-6 lg:grid-cols-2">
        {/* Quick Actions */}
        <div className="bg-card border rounded-lg p-6">
          <h2 className="text-lg font-semibold mb-4">{t('quickActions')}</h2>
          <div className="grid gap-3">
            <a
              href="/solve"
              className="flex items-center gap-3 p-4 rounded-lg border hover:bg-muted transition-colors"
            >
              <Image className="h-5 w-5 text-primary" />
              <div>
                <p className="font-medium">Giải Captcha</p>
                <p className="text-sm text-muted-foreground">Tải ảnh và giải captcha ngay</p>
              </div>
            </a>
            <a
              href="/api-keys"
              className="flex items-center gap-3 p-4 rounded-lg border hover:bg-muted transition-colors"
            >
              <Key className="h-5 w-5 text-primary" />
              <div>
                <p className="font-medium">Quản lý API Keys</p>
                <p className="text-sm text-muted-foreground">Tạo và quản lý API keys</p>
              </div>
            </a>
            <a
              href="/training"
              className="flex items-center gap-3 p-4 rounded-lg border hover:bg-muted transition-colors"
            >
              <TrendingUp className="h-5 w-5 text-primary" />
              <div>
                <p className="font-medium">Huấn luyện Model</p>
                <p className="text-sm text-muted-foreground">Tạo model mới từ dữ liệu của bạn</p>
              </div>
            </a>
          </div>
        </div>

        {/* API Keys Summary */}
        <div className="bg-card border rounded-lg p-6">
          <h2 className="text-lg font-semibold mb-4">API Keys</h2>
          {apiKeys?.data && apiKeys.data.length > 0 ? (
            <div className="space-y-3">
              {apiKeys.data.slice(0, 3).map((key) => (
                <div
                  key={key.id}
                  className="flex items-center justify-between p-3 rounded-lg border"
                >
                  <div>
                    <p className="font-medium">{key.name}</p>
                    <p className="text-sm text-muted-foreground font-mono">
                      {key.key_prefix}...
                    </p>
                  </div>
                  <div className="text-right">
                    <p className="text-sm">
                      {key.total_requests.toLocaleString()} requests
                    </p>
                    <p className={`text-sm ${key.is_active ? 'text-green-500' : 'text-red-500'}`}>
                      {key.is_active ? 'Active' : 'Inactive'}
                    </p>
                  </div>
                </div>
              ))}
              {apiKeys.data.length > 3 && (
                <a href="/api-keys" className="block text-center text-sm text-primary hover:underline">
                  Xem tất cả ({apiKeys.data.length})
                </a>
              )}
            </div>
          ) : (
            <div className="text-center py-8 text-muted-foreground">
              <Key className="h-12 w-12 mx-auto mb-3 opacity-50" />
              <p>Chưa có API key nào</p>
              <a href="/api-keys" className="text-primary hover:underline text-sm">
                Tạo API key đầu tiên
              </a>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}