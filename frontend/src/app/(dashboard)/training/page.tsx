'use client';

import { useState } from 'react';
import { useTranslations } from 'next-intl';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { 
  TrendingUp, 
  Play, 
  Pause, 
  X, 
  Clock, 
  CheckCircle,
  XCircle,
  Loader2,
  Plus
} from 'lucide-react';
import { toast } from 'sonner';

import { Button } from '@/components/ui/button';
import { trainingApi, captchaApi } from '@/services/api';

type TrainingJob = {
  id: number;
  name: string;
  status: 'pending' | 'running' | 'completed' | 'failed' | 'cancelled';
  model_type: string;
  progress: number;
  current_epoch?: number;
  total_epochs?: number;
  started_at?: string;
  completed_at?: string;
  created_at: string;
};

export default function TrainingPage() {
  const t = useTranslations('training');
  const tCommon = useTranslations('common');
  const queryClient = useQueryClient();

  const [showNewJobModal, setShowNewJobModal] = useState(false);

  // Form state for new job
  const [newJob, setNewJob] = useState({
    name: '',
    model_type: 'cnn',
    epochs: 100,
    batch_size: 32,
    learning_rate: 0.001,
    validation_split: 0.2,
    dataset_path: '',
  });

  // Fetch training jobs
  const { data: jobsData, isLoading } = useQuery({
    queryKey: ['training-jobs'],
    queryFn: async () => {
      // Mock data for now
      return { data: [] as TrainingJob[] };
    },
    refetchInterval: 5000, // Poll every 5 seconds
  });

  const jobs = jobsData?.data || [];

  // Start training mutation
  const startMutation = useMutation({
    mutationFn: trainingApi.start,
    onSuccess: () => {
      toast.success(t('startSuccess'));
      queryClient.invalidateQueries({ queryKey: ['training-jobs'] });
      setShowNewJobModal(false);
      setNewJob({
        name: '',
        model_type: 'cnn',
        epochs: 100,
        batch_size: 32,
        learning_rate: 0.001,
        validation_split: 0.2,
        dataset_path: '',
      });
    },
    onError: (error: any) => {
      toast.error(error.response?.data?.message || t('startError'));
    },
  });

  const handleStartTraining = () => {
    if (!newJob.name) {
      toast.error('Vui lòng nhập tên job');
      return;
    }

    startMutation.mutate({
      name: newJob.name,
      model_type: newJob.model_type,
      config: {
        epochs: newJob.epochs,
        batch_size: newJob.batch_size,
        learning_rate: newJob.learning_rate,
        validation_split: newJob.validation_split,
      },
      dataset_path: newJob.dataset_path || undefined,
    });
  };

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'pending':
        return <Clock className="h-5 w-5 text-yellow-500" />;
      case 'running':
        return <Loader2 className="h-5 w-5 text-blue-500 animate-spin" />;
      case 'completed':
        return <CheckCircle className="h-5 w-5 text-green-500" />;
      case 'failed':
        return <XCircle className="h-5 w-5 text-red-500" />;
      case 'cancelled':
        return <X className="h-5 w-5 text-gray-500" />;
      default:
        return <Clock className="h-5 w-5 text-gray-500" />;
    }
  };

  const getStatusText = (status: string) => {
    switch (status) {
      case 'pending':
        return t('pending');
      case 'running':
        return t('running');
      case 'completed':
        return t('completed');
      case 'failed':
        return t('failed');
      case 'cancelled':
        return t('cancelled');
      default:
        return status;
    }
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold">{t('title')}</h1>
          <p className="text-muted-foreground mt-1">
            Huấn luyện model mới từ dữ liệu của bạn
          </p>
        </div>
        <Button onClick={() => setShowNewJobModal(true)}>
          <Plus className="mr-2 h-4 w-4" />
          {t('startTraining')}
        </Button>
      </div>

      {/* Training Jobs List */}
      {isLoading ? (
        <div className="flex items-center justify-center py-12">
          <Loader2 className="h-8 w-8 animate-spin text-primary" />
        </div>
      ) : jobs.length === 0 ? (
        <div className="text-center py-12 bg-card border rounded-lg">
          <TrendingUp className="h-12 w-12 mx-auto mb-4 text-muted-foreground opacity-50" />
          <p className="text-muted-foreground">Chưa có job huấn luyện nào</p>
          <Button 
            variant="outline" 
            className="mt-4"
            onClick={() => setShowNewJobModal(true)}
          >
            <Plus className="mr-2 h-4 w-4" />
            Bắt đầu huấn luyện
          </Button>
        </div>
      ) : (
        <div className="space-y-4">
          {jobs.map((job) => (
            <div
              key={job.id}
              className="bg-card border rounded-lg p-5"
            >
              <div className="flex items-start justify-between mb-4">
                <div className="flex items-center gap-3">
                  {getStatusIcon(job.status)}
                  <div>
                    <h3 className="font-semibold">{job.name}</h3>
                    <p className="text-sm text-muted-foreground">
                      {job.model_type.toUpperCase()} • {getStatusText(job.status)}
                    </p>
                  </div>
                </div>
                <div className="text-right text-sm">
                  <p className="text-muted-foreground">
                    {new Date(job.created_at).toLocaleDateString('vi-VN')}
                  </p>
                  {job.current_epoch && job.total_epochs && (
                    <p className="font-medium">
                      Epoch {job.current_epoch}/{job.total_epochs}
                    </p>
                  )}
                </div>
              </div>

              {/* Progress bar */}
              {(job.status === 'running' || job.status === 'completed') && (
                <div className="mb-4">
                  <div className="flex justify-between text-sm mb-1">
                    <span className="text-muted-foreground">{t('progress')}</span>
                    <span className="font-medium">{job.progress.toFixed(1)}%</span>
                  </div>
                  <div className="h-2 bg-muted rounded-full overflow-hidden">
                    <div 
                      className={`h-full rounded-full transition-all ${
                        job.status === 'completed' ? 'bg-green-500' : 'bg-primary'
                      }`}
                      style={{ width: `${job.progress}%` }}
                    />
                  </div>
                </div>
              )}

              {/* Actions */}
              {job.status === 'running' && (
                <div className="flex gap-2">
                  <Button variant="outline" size="sm">
                    <Pause className="mr-1 h-3 w-3" />
                    Tạm dừng
                  </Button>
                  <Button variant="outline" size="sm" className="text-destructive">
                    <X className="mr-1 h-3 w-3" />
                    Hủy
                  </Button>
                </div>
              )}
            </div>
          ))}
        </div>
      )}

      {/* New Job Modal */}
      {showNewJobModal && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
          <div className="bg-card border rounded-lg p-6 w-full max-w-lg mx-4 max-h-[90vh] overflow-y-auto">
            <h2 className="text-xl font-semibold mb-4">{t('startTraining')}</h2>
            
            <div className="space-y-4">
              <div>
                <label className="block text-sm font-medium mb-1">{t('jobName')}</label>
                <input
                  type="text"
                  value={newJob.name}
                  onChange={(e) => setNewJob({ ...newJob, name: e.target.value })}
                  className="w-full px-3 py-2 bg-background border rounded-lg"
                  placeholder="my-training-job"
                />
              </div>

              <div>
                <label className="block text-sm font-medium mb-1">{t('modelType')}</label>
                <select
                  value={newJob.model_type}
                  onChange={(e) => setNewJob({ ...newJob, model_type: e.target.value })}
                  className="w-full px-3 py-2 bg-background border rounded-lg"
                >
                  <option value="ocr">OCR</option>
                  <option value="cnn">CNN</option>
                  <option value="rnn">RNN</option>
                  <option value="transformer">Transformer</option>
                </select>
              </div>

              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="block text-sm font-medium mb-1">{t('epochs')}</label>
                  <input
                    type="number"
                    value={newJob.epochs}
                    onChange={(e) => setNewJob({ ...newJob, epochs: parseInt(e.target.value) })}
                    className="w-full px-3 py-2 bg-background border rounded-lg"
                    min={1}
                    max={1000}
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium mb-1">{t('batchSize')}</label>
                  <input
                    type="number"
                    value={newJob.batch_size}
                    onChange={(e) => setNewJob({ ...newJob, batch_size: parseInt(e.target.value) })}
                    className="w-full px-3 py-2 bg-background border rounded-lg"
                    min={1}
                    max={256}
                  />
                </div>
              </div>

              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="block text-sm font-medium mb-1">{t('learningRate')}</label>
                  <input
                    type="number"
                    value={newJob.learning_rate}
                    onChange={(e) => setNewJob({ ...newJob, learning_rate: parseFloat(e.target.value) })}
                    className="w-full px-3 py-2 bg-background border rounded-lg"
                    step={0.0001}
                    min={0.0001}
                    max={1}
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium mb-1">{t('validationSplit')}</label>
                  <input
                    type="number"
                    value={newJob.validation_split}
                    onChange={(e) => setNewJob({ ...newJob, validation_split: parseFloat(e.target.value) })}
                    className="w-full px-3 py-2 bg-background border rounded-lg"
                    step={0.1}
                    min={0.1}
                    max={0.5}
                  />
                </div>
              </div>

              <div>
                <label className="block text-sm font-medium mb-1">{t('datasetPath')}</label>
                <input
                  type="text"
                  value={newJob.dataset_path}
                  onChange={(e) => setNewJob({ ...newJob, dataset_path: e.target.value })}
                  className="w-full px-3 py-2 bg-background border rounded-lg"
                  placeholder="/path/to/dataset (optional)"
                />
                <p className="text-xs text-muted-foreground mt-1">
                  Để trống để sử dụng dataset mặc định
                </p>
              </div>
            </div>

            <div className="flex gap-3 mt-6">
              <Button
                variant="outline"
                className="flex-1"
                onClick={() => setShowNewJobModal(false)}
              >
                {tCommon('cancel')}
              </Button>
              <Button
                className="flex-1"
                onClick={handleStartTraining}
                disabled={startMutation.isPending}
              >
                {startMutation.isPending ? (
                  <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                ) : (
                  <Play className="mr-2 h-4 w-4" />
                )}
                {t('startTraining')}
              </Button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}