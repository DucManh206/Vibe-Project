'use client';

import { useState } from 'react';
import { useTranslations } from 'next-intl';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { 
  Key, 
  Plus, 
  Trash2, 
  Copy, 
  Check, 
  AlertTriangle,
  Loader2,
  X
} from 'lucide-react';
import { toast } from 'sonner';

import { Button } from '@/components/ui/button';
import { apiKeysApi, CreateApiKeyRequest } from '@/services/api';

export default function ApiKeysPage() {
  const t = useTranslations('apiKeys');
  const tCommon = useTranslations('common');
  const queryClient = useQueryClient();

  const [showCreateModal, setShowCreateModal] = useState(false);
  const [newKeyResult, setNewKeyResult] = useState<{ key: string; id: number } | null>(null);
  const [deleteConfirm, setDeleteConfirm] = useState<number | null>(null);
  const [copiedId, setCopiedId] = useState<string | null>(null);

  // Form state
  const [newKey, setNewKey] = useState<CreateApiKeyRequest>({
    name: '',
    rate_limit: 100,
    expires_in: 0,
    scopes: [],
  });

  // Fetch API keys
  const { data: apiKeysData, isLoading } = useQuery({
    queryKey: ['apiKeys'],
    queryFn: () => apiKeysApi.list(),
  });

  const apiKeys = apiKeysData?.data || [];

  // Create mutation
  const createMutation = useMutation({
    mutationFn: apiKeysApi.create,
    onSuccess: (response) => {
      setNewKeyResult({ key: response.data.key, id: response.data.id });
      queryClient.invalidateQueries({ queryKey: ['apiKeys'] });
      toast.success(t('createSuccess'));
    },
    onError: (error: any) => {
      toast.error(error.response?.data?.message || t('createError'));
    },
  });

  // Delete mutation
  const deleteMutation = useMutation({
    mutationFn: apiKeysApi.delete,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['apiKeys'] });
      toast.success(t('deleteSuccess'));
      setDeleteConfirm(null);
    },
    onError: (error: any) => {
      toast.error(error.response?.data?.message || 'Xóa API key thất bại');
    },
  });

  const handleCreate = () => {
    if (!newKey.name) {
      toast.error('Vui lòng nhập tên API key');
      return;
    }
    createMutation.mutate(newKey);
  };

  const handleCopy = (text: string, id: string) => {
    navigator.clipboard.writeText(text);
    setCopiedId(id);
    setTimeout(() => setCopiedId(null), 2000);
    toast.success(t('keyCopied'));
  };

  const closeCreateModal = () => {
    setShowCreateModal(false);
    setNewKeyResult(null);
    setNewKey({ name: '', rate_limit: 100, expires_in: 0, scopes: [] });
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold">{t('title')}</h1>
          <p className="text-muted-foreground mt-1">
            Quản lý API keys để tích hợp vào ứng dụng của bạn
          </p>
        </div>
        <Button onClick={() => setShowCreateModal(true)}>
          <Plus className="mr-2 h-4 w-4" />
          {t('create')}
        </Button>
      </div>

      {/* API Keys List */}
      {isLoading ? (
        <div className="flex items-center justify-center py-12">
          <Loader2 className="h-8 w-8 animate-spin text-primary" />
        </div>
      ) : apiKeys.length === 0 ? (
        <div className="text-center py-12 bg-card border rounded-lg">
          <Key className="h-12 w-12 mx-auto mb-4 text-muted-foreground opacity-50" />
          <p className="text-muted-foreground">Chưa có API key nào</p>
          <Button 
            variant="outline" 
            className="mt-4"
            onClick={() => setShowCreateModal(true)}
          >
            <Plus className="mr-2 h-4 w-4" />
            Tạo API key đầu tiên
          </Button>
        </div>
      ) : (
        <div className="space-y-4">
          {apiKeys.map((apiKey) => (
            <div
              key={apiKey.id}
              className="bg-card border rounded-lg p-5"
            >
              <div className="flex items-start justify-between">
                <div className="flex items-center gap-3">
                  <div className="p-2 bg-primary/10 rounded-lg">
                    <Key className="h-5 w-5 text-primary" />
                  </div>
                  <div>
                    <h3 className="font-semibold">{apiKey.name}</h3>
                    <div className="flex items-center gap-2 mt-1">
                      <code className="text-sm bg-muted px-2 py-0.5 rounded font-mono">
                        {apiKey.key_prefix}...
                      </code>
                      <span className={`text-xs px-2 py-0.5 rounded-full ${
                        apiKey.is_active 
                          ? 'bg-green-500/10 text-green-600'
                          : 'bg-red-500/10 text-red-600'
                      }`}>
                        {apiKey.is_active ? t('active') : t('inactive')}
                      </span>
                    </div>
                  </div>
                </div>

                {/* Actions */}
                <div className="flex items-center gap-2">
                  {deleteConfirm === apiKey.id ? (
                    <>
                      <Button
                        variant="destructive"
                        size="sm"
                        onClick={() => deleteMutation.mutate(apiKey.id)}
                        disabled={deleteMutation.isPending}
                      >
                        {deleteMutation.isPending ? (
                          <Loader2 className="h-4 w-4 animate-spin" />
                        ) : (
                          <Check className="h-4 w-4" />
                        )}
                      </Button>
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={() => setDeleteConfirm(null)}
                      >
                        <X className="h-4 w-4" />
                      </Button>
                    </>
                  ) : (
                    <Button
                      variant="ghost"
                      size="sm"
                      onClick={() => setDeleteConfirm(apiKey.id)}
                    >
                      <Trash2 className="h-4 w-4 text-destructive" />
                    </Button>
                  )}
                </div>
              </div>

              {/* Stats */}
              <div className="grid grid-cols-3 gap-4 mt-4 pt-4 border-t">
                <div>
                  <p className="text-sm text-muted-foreground">{t('rateLimit')}</p>
                  <p className="font-medium">{apiKey.rate_limit} req/min</p>
                </div>
                <div>
                  <p className="text-sm text-muted-foreground">{t('totalRequests')}</p>
                  <p className="font-medium">{apiKey.total_requests.toLocaleString()}</p>
                </div>
                <div>
                  <p className="text-sm text-muted-foreground">{t('lastUsed')}</p>
                  <p className="font-medium">
                    {apiKey.last_used_at 
                      ? new Date(apiKey.last_used_at).toLocaleDateString('vi-VN')
                      : 'Chưa sử dụng'
                    }
                  </p>
                </div>
              </div>

              {/* Expiry warning */}
              {apiKey.expires_at && new Date(apiKey.expires_at) < new Date(Date.now() + 7 * 24 * 60 * 60 * 1000) && (
                <div className="flex items-center gap-2 mt-4 p-3 bg-yellow-500/10 text-yellow-700 rounded-lg">
                  <AlertTriangle className="h-4 w-4" />
                  <span className="text-sm">
                    Key sẽ hết hạn vào {new Date(apiKey.expires_at).toLocaleDateString('vi-VN')}
                  </span>
                </div>
              )}
            </div>
          ))}
        </div>
      )}

      {/* Create Modal */}
      {showCreateModal && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
          <div className="bg-card border rounded-lg p-6 w-full max-w-md mx-4">
            {newKeyResult ? (
              // Show new key
              <>
                <div className="flex items-center gap-2 mb-4">
                  <Check className="h-6 w-6 text-green-500" />
                  <h2 className="text-xl font-semibold">{t('createSuccess')}</h2>
                </div>

                <div className="bg-yellow-500/10 border border-yellow-500/20 rounded-lg p-4 mb-4">
                  <div className="flex items-start gap-2">
                    <AlertTriangle className="h-5 w-5 text-yellow-600 shrink-0 mt-0.5" />
                    <p className="text-sm text-yellow-700">{t('keyWarning')}</p>
                  </div>
                </div>

                <div className="relative">
                  <code className="block w-full p-3 bg-muted rounded-lg text-sm font-mono break-all pr-12">
                    {newKeyResult.key}
                  </code>
                  <Button
                    variant="ghost"
                    size="sm"
                    className="absolute right-2 top-1/2 -translate-y-1/2"
                    onClick={() => handleCopy(newKeyResult.key, 'new-key')}
                  >
                    {copiedId === 'new-key' ? (
                      <Check className="h-4 w-4 text-green-500" />
                    ) : (
                      <Copy className="h-4 w-4" />
                    )}
                  </Button>
                </div>

                <Button className="w-full mt-6" onClick={closeCreateModal}>
                  {tCommon('close')}
                </Button>
              </>
            ) : (
              // Create form
              <>
                <h2 className="text-xl font-semibold mb-4">{t('create')}</h2>
                
                <div className="space-y-4">
                  <div>
                    <label className="block text-sm font-medium mb-1">{t('name')}</label>
                    <input
                      type="text"
                      value={newKey.name}
                      onChange={(e) => setNewKey({ ...newKey, name: e.target.value })}
                      className="w-full px-3 py-2 bg-background border rounded-lg"
                      placeholder="My API Key"
                    />
                  </div>

                  <div>
                    <label className="block text-sm font-medium mb-1">{t('rateLimit')}</label>
                    <input
                      type="number"
                      value={newKey.rate_limit}
                      onChange={(e) => setNewKey({ ...newKey, rate_limit: parseInt(e.target.value) })}
                      className="w-full px-3 py-2 bg-background border rounded-lg"
                      min={1}
                      max={1000}
                    />
                    <p className="text-xs text-muted-foreground mt-1">Số requests tối đa mỗi phút</p>
                  </div>

                  <div>
                    <label className="block text-sm font-medium mb-1">{t('expiresInDays')}</label>
                    <select
                      value={newKey.expires_in}
                      onChange={(e) => setNewKey({ ...newKey, expires_in: parseInt(e.target.value) })}
                      className="w-full px-3 py-2 bg-background border rounded-lg"
                    >
                      <option value={0}>{t('noExpiry')}</option>
                      <option value={30}>30 ngày</option>
                      <option value={90}>90 ngày</option>
                      <option value={180}>180 ngày</option>
                      <option value={365}>1 năm</option>
                    </select>
                  </div>
                </div>

                <div className="flex gap-3 mt-6">
                  <Button
                    variant="outline"
                    className="flex-1"
                    onClick={closeCreateModal}
                  >
                    {tCommon('cancel')}
                  </Button>
                  <Button
                    className="flex-1"
                    onClick={handleCreate}
                    disabled={createMutation.isPending}
                  >
                    {createMutation.isPending ? (
                      <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                    ) : (
                      <Plus className="mr-2 h-4 w-4" />
                    )}
                    {tCommon('create')}
                  </Button>
                </div>
              </>
            )}
          </div>
        </div>
      )}
    </div>
  );
}