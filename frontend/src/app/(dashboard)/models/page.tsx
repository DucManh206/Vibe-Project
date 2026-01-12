'use client';

import { useState } from 'react';
import { useTranslations } from 'next-intl';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { 
  Cpu, 
  Plus, 
  Trash2, 
  Star, 
  StarOff, 
  Check, 
  X,
  Upload,
  Loader2
} from 'lucide-react';
import { toast } from 'sonner';

import { Button } from '@/components/ui/button';
import { captchaApi } from '@/services/api';

export default function ModelsPage() {
  const t = useTranslations('models');
  const tCommon = useTranslations('common');
  const queryClient = useQueryClient();

  const [showUploadModal, setShowUploadModal] = useState(false);
  const [deleteConfirm, setDeleteConfirm] = useState<number | null>(null);

  // Fetch models
  const { data: modelsData, isLoading } = useQuery({
    queryKey: ['models'],
    queryFn: () => captchaApi.getModels(),
  });

  const models = modelsData?.data || [];

  // Upload form state
  const [uploadForm, setUploadForm] = useState({
    name: '',
    type: 'cnn',
    version: '1.0.0',
    description: '',
  });

  // Upload mutation
  const uploadMutation = useMutation({
    mutationFn: async (formData: FormData) => {
      return captchaApi.uploadModel(formData);
    },
    onSuccess: () => {
      toast.success(t('uploadSuccess'));
      queryClient.invalidateQueries({ queryKey: ['models'] });
      setShowUploadModal(false);
      setUploadForm({ name: '', type: 'cnn', version: '1.0.0', description: '' });
    },
    onError: (error: any) => {
      toast.error(error.response?.data?.message || t('uploadError'));
    },
  });

  const handleUpload = () => {
    if (!uploadForm.name) {
      toast.error('Vui lòng nhập tên model');
      return;
    }

    const formData = new FormData();
    formData.append('name', uploadForm.name);
    formData.append('model_type', uploadForm.type);
    formData.append('version', uploadForm.version);
    formData.append('description', uploadForm.description);

    uploadMutation.mutate(formData);
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold">{t('title')}</h1>
          <p className="text-muted-foreground mt-1">
            Quản lý các model giải captcha
          </p>
        </div>
        <Button onClick={() => setShowUploadModal(true)}>
          <Plus className="mr-2 h-4 w-4" />
          {t('upload')}
        </Button>
      </div>

      {/* Models Grid */}
      {isLoading ? (
        <div className="flex items-center justify-center py-12">
          <Loader2 className="h-8 w-8 animate-spin text-primary" />
        </div>
      ) : models.length === 0 ? (
        <div className="text-center py-12 bg-card border rounded-lg">
          <Cpu className="h-12 w-12 mx-auto mb-4 text-muted-foreground opacity-50" />
          <p className="text-muted-foreground">Chưa có model nào</p>
          <Button 
            variant="outline" 
            className="mt-4"
            onClick={() => setShowUploadModal(true)}
          >
            <Plus className="mr-2 h-4 w-4" />
            Thêm model đầu tiên
          </Button>
        </div>
      ) : (
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
          {models.map((model) => (
            <div
              key={model.id}
              className="bg-card border rounded-lg p-5 relative"
            >
              {/* Status badges */}
              <div className="absolute top-3 right-3 flex gap-2">
                {model.is_default && (
                  <span className="px-2 py-1 bg-yellow-500/10 text-yellow-600 text-xs rounded-full flex items-center gap-1">
                    <Star className="h-3 w-3" />
                    {t('default')}
                  </span>
                )}
                <span className={`px-2 py-1 text-xs rounded-full ${
                  model.is_active 
                    ? 'bg-green-500/10 text-green-600'
                    : 'bg-red-500/10 text-red-600'
                }`}>
                  {model.is_active ? t('active') : t('inactive')}
                </span>
              </div>

              {/* Model info */}
              <div className="flex items-start gap-3 mb-4">
                <div className="p-2 bg-primary/10 rounded-lg">
                  <Cpu className="h-6 w-6 text-primary" />
                </div>
                <div>
                  <h3 className="font-semibold">{model.name}</h3>
                  <p className="text-sm text-muted-foreground">
                    {model.type.toUpperCase()} • v{model.version}
                  </p>
                </div>
              </div>

              {/* Stats */}
              {model.accuracy && (
                <div className="mb-4">
                  <div className="flex justify-between text-sm mb-1">
                    <span className="text-muted-foreground">{t('accuracy')}</span>
                    <span className="font-medium">{(model.accuracy * 100).toFixed(1)}%</span>
                  </div>
                  <div className="h-2 bg-muted rounded-full overflow-hidden">
                    <div 
                      className="h-full bg-primary rounded-full"
                      style={{ width: `${model.accuracy * 100}%` }}
                    />
                  </div>
                </div>
              )}

              {/* Description */}
              {model.description && (
                <p className="text-sm text-muted-foreground mb-4 line-clamp-2">
                  {model.description}
                </p>
              )}

              {/* Actions */}
              <div className="flex gap-2 pt-4 border-t">
                {!model.is_default && model.is_active && (
                  <Button variant="outline" size="sm" className="flex-1">
                    <Star className="mr-1 h-3 w-3" />
                    Set Default
                  </Button>
                )}
                {deleteConfirm === model.id ? (
                  <div className="flex gap-2 flex-1">
                    <Button
                      variant="destructive"
                      size="sm"
                      className="flex-1"
                      onClick={() => {
                        // Delete logic here
                        setDeleteConfirm(null);
                        toast.success(t('deleteSuccess'));
                      }}
                    >
                      <Check className="h-4 w-4" />
                    </Button>
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() => setDeleteConfirm(null)}
                    >
                      <X className="h-4 w-4" />
                    </Button>
                  </div>
                ) : (
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => setDeleteConfirm(model.id)}
                  >
                    <Trash2 className="h-4 w-4 text-destructive" />
                  </Button>
                )}
              </div>

              {/* Created date */}
              <p className="text-xs text-muted-foreground mt-3">
                {t('createdAt')}: {new Date(model.created_at).toLocaleDateString('vi-VN')}
              </p>
            </div>
          ))}
        </div>
      )}

      {/* Upload Modal */}
      {showUploadModal && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
          <div className="bg-card border rounded-lg p-6 w-full max-w-md mx-4">
            <h2 className="text-xl font-semibold mb-4">{t('upload')}</h2>
            
            <div className="space-y-4">
              <div>
                <label className="block text-sm font-medium mb-1">{t('name')}</label>
                <input
                  type="text"
                  value={uploadForm.name}
                  onChange={(e) => setUploadForm({ ...uploadForm, name: e.target.value })}
                  className="w-full px-3 py-2 bg-background border rounded-lg"
                  placeholder="my-captcha-model"
                />
              </div>

              <div>
                <label className="block text-sm font-medium mb-1">{t('type')}</label>
                <select
                  value={uploadForm.type}
                  onChange={(e) => setUploadForm({ ...uploadForm, type: e.target.value })}
                  className="w-full px-3 py-2 bg-background border rounded-lg"
                >
                  <option value="ocr">{t('typeOcr')}</option>
                  <option value="cnn">{t('typeCnn')}</option>
                  <option value="rnn">{t('typeRnn')}</option>
                  <option value="transformer">{t('typeTransformer')}</option>
                  <option value="ensemble">{t('typeEnsemble')}</option>
                </select>
              </div>

              <div>
                <label className="block text-sm font-medium mb-1">{t('version')}</label>
                <input
                  type="text"
                  value={uploadForm.version}
                  onChange={(e) => setUploadForm({ ...uploadForm, version: e.target.value })}
                  className="w-full px-3 py-2 bg-background border rounded-lg"
                  placeholder="1.0.0"
                />
              </div>

              <div>
                <label className="block text-sm font-medium mb-1">{t('description')}</label>
                <textarea
                  value={uploadForm.description}
                  onChange={(e) => setUploadForm({ ...uploadForm, description: e.target.value })}
                  className="w-full px-3 py-2 bg-background border rounded-lg h-20 resize-none"
                  placeholder="Mô tả model..."
                />
              </div>
            </div>

            <div className="flex gap-3 mt-6">
              <Button
                variant="outline"
                className="flex-1"
                onClick={() => setShowUploadModal(false)}
              >
                {tCommon('cancel')}
              </Button>
              <Button
                className="flex-1"
                onClick={handleUpload}
                disabled={uploadMutation.isPending}
              >
                {uploadMutation.isPending ? (
                  <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                ) : (
                  <Upload className="mr-2 h-4 w-4" />
                )}
                {t('upload')}
              </Button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}