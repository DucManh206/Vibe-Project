'use client';

import { useState, useCallback } from 'react';
import { useTranslations } from 'next-intl';
import { useDropzone } from 'react-dropzone';
import { useMutation, useQuery } from '@tanstack/react-query';
import { Upload, Copy, Check, Loader2, Image as ImageIcon, X } from 'lucide-react';
import { toast } from 'sonner';

import { Button } from '@/components/ui/button';
import { captchaApi } from '@/services/api';

export default function SolvePage() {
  const t = useTranslations('captcha');
  const tCommon = useTranslations('common');

  const [imageBase64, setImageBase64] = useState<string | null>(null);
  const [imagePreview, setImagePreview] = useState<string | null>(null);
  const [selectedModel, setSelectedModel] = useState<string>('');
  const [copied, setCopied] = useState(false);

  // Fetch available models
  const { data: modelsData } = useQuery({
    queryKey: ['models'],
    queryFn: () => captchaApi.getModels(),
  });

  const models = modelsData?.data || [];

  // Solve mutation
  const solveMutation = useMutation({
    mutationFn: captchaApi.solve,
    onSuccess: () => {
      toast.success(tCommon('success'));
    },
    onError: (error: any) => {
      toast.error(error.response?.data?.message || t('solveError'));
    },
  });

  // Handle file drop
  const onDrop = useCallback((acceptedFiles: File[]) => {
    const file = acceptedFiles[0];
    if (file) {
      // Check file size (max 10MB)
      if (file.size > 10 * 1024 * 1024) {
        toast.error(t('imageTooLarge'));
        return;
      }

      const reader = new FileReader();
      reader.onload = () => {
        const result = reader.result as string;
        setImagePreview(result);
        // Remove data URL prefix for API
        setImageBase64(result.split(',')[1]);
      };
      reader.readAsDataURL(file);
    }
  }, [t]);

  const { getRootProps, getInputProps, isDragActive } = useDropzone({
    onDrop,
    accept: {
      'image/*': ['.png', '.jpg', '.jpeg', '.gif', '.webp'],
    },
    maxFiles: 1,
  });

  // Handle paste from clipboard
  const handlePaste = (e: React.ClipboardEvent) => {
    const items = e.clipboardData.items;
    for (let i = 0; i < items.length; i++) {
      if (items[i].type.indexOf('image') !== -1) {
        const file = items[i].getAsFile();
        if (file) {
          onDrop([file]);
        }
        break;
      }
    }
  };

  // Handle base64 input
  const handleBase64Input = (e: React.ChangeEvent<HTMLTextAreaElement>) => {
    const value = e.target.value.trim();
    if (value) {
      // Check if it's a data URL or raw base64
      if (value.startsWith('data:')) {
        setImagePreview(value);
        setImageBase64(value.split(',')[1]);
      } else {
        setImagePreview(`data:image/png;base64,${value}`);
        setImageBase64(value);
      }
    } else {
      setImagePreview(null);
      setImageBase64(null);
    }
  };

  // Clear image
  const clearImage = () => {
    setImageBase64(null);
    setImagePreview(null);
    solveMutation.reset();
  };

  // Solve captcha
  const handleSolve = () => {
    if (!imageBase64) {
      toast.error(t('noImage'));
      return;
    }

    solveMutation.mutate({
      image_base64: imageBase64,
      model: selectedModel || undefined,
    });
  };

  // Copy result
  const copyResult = () => {
    if (solveMutation.data?.data?.text) {
      navigator.clipboard.writeText(solveMutation.data.data.text);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
      toast.success(t('copied'));
    }
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div>
        <h1 className="text-3xl font-bold">{t('title')}</h1>
        <p className="text-muted-foreground mt-1">
          Tải ảnh captcha lên để nhận kết quả giải
        </p>
      </div>

      <div className="grid gap-6 lg:grid-cols-2">
        {/* Left Column - Image Input */}
        <div className="space-y-4">
          {/* Model Selection */}
          <div className="bg-card border rounded-lg p-4">
            <label className="block text-sm font-medium mb-2">
              {t('selectModel')}
            </label>
            <select
              value={selectedModel}
              onChange={(e) => setSelectedModel(e.target.value)}
              className="w-full px-3 py-2 bg-background border rounded-lg focus:ring-2 focus:ring-primary"
            >
              <option value="">Tự động chọn</option>
              {models.map((model) => (
                <option key={model.id} value={model.name}>
                  {model.name} ({model.type})
                  {model.accuracy && ` - ${(model.accuracy * 100).toFixed(1)}%`}
                </option>
              ))}
            </select>
          </div>

          {/* Dropzone */}
          <div
            {...getRootProps()}
            onPaste={handlePaste}
            className={`
              relative border-2 border-dashed rounded-lg p-8 text-center cursor-pointer
              transition-colors
              ${isDragActive ? 'border-primary bg-primary/5' : 'border-muted-foreground/25 hover:border-primary'}
            `}
          >
            <input {...getInputProps()} />
            
            {imagePreview ? (
              <div className="relative">
                <img
                  src={imagePreview}
                  alt="Captcha preview"
                  className="max-h-48 mx-auto rounded-lg"
                />
                <button
                  onClick={(e) => {
                    e.stopPropagation();
                    clearImage();
                  }}
                  className="absolute top-2 right-2 p-1 bg-destructive text-destructive-foreground rounded-full"
                >
                  <X className="h-4 w-4" />
                </button>
              </div>
            ) : (
              <div className="space-y-4">
                <div className="mx-auto w-16 h-16 bg-primary/10 rounded-full flex items-center justify-center">
                  <Upload className="h-8 w-8 text-primary" />
                </div>
                <div>
                  <p className="font-medium">{t('uploadImage')}</p>
                  <p className="text-sm text-muted-foreground mt-1">
                    {t('dropImage')}
                  </p>
                  <p className="text-xs text-muted-foreground mt-2">
                    PNG, JPG, GIF, WEBP (max 10MB)
                  </p>
                </div>
              </div>
            )}
          </div>

          {/* Base64 Input */}
          <div className="bg-card border rounded-lg p-4">
            <label className="block text-sm font-medium mb-2">
              {t('pasteBase64')}
            </label>
            <textarea
              placeholder="data:image/png;base64,... hoặc raw base64"
              onChange={handleBase64Input}
              className="w-full h-24 px-3 py-2 bg-background border rounded-lg text-xs font-mono resize-none focus:ring-2 focus:ring-primary"
            />
          </div>

          {/* Solve Button */}
          <Button
            onClick={handleSolve}
            disabled={!imageBase64 || solveMutation.isPending}
            className="w-full"
            size="lg"
          >
            {solveMutation.isPending ? (
              <>
                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                {t('solving')}
              </>
            ) : (
              <>
                <ImageIcon className="mr-2 h-4 w-4" />
                {t('solve')}
              </>
            )}
          </Button>
        </div>

        {/* Right Column - Results */}
        <div className="space-y-4">
          <div className="bg-card border rounded-lg p-6">
            <h2 className="text-lg font-semibold mb-4">{t('result')}</h2>
            
            {solveMutation.data?.data ? (
              <div className="space-y-4">
                {/* Result Text */}
                <div className="relative">
                  <div className="p-4 bg-muted rounded-lg">
                    <p className="text-3xl font-mono font-bold text-center tracking-wider">
                      {solveMutation.data.data.text}
                    </p>
                  </div>
                  <Button
                    variant="ghost"
                    size="sm"
                    className="absolute top-2 right-2"
                    onClick={copyResult}
                  >
                    {copied ? (
                      <Check className="h-4 w-4 text-green-500" />
                    ) : (
                      <Copy className="h-4 w-4" />
                    )}
                  </Button>
                </div>

                {/* Details */}
                <div className="grid grid-cols-2 gap-4 text-sm">
                  <div className="p-3 bg-muted/50 rounded-lg">
                    <p className="text-muted-foreground">{t('confidence')}</p>
                    <p className="text-lg font-semibold">
                      {(solveMutation.data.data.confidence * 100).toFixed(1)}%
                    </p>
                  </div>
                  <div className="p-3 bg-muted/50 rounded-lg">
                    <p className="text-muted-foreground">{t('processingTime')}</p>
                    <p className="text-lg font-semibold">
                      {solveMutation.data.data.processing_time_ms}ms
                    </p>
                  </div>
                </div>

                {/* Model Used */}
                <div className="text-sm text-muted-foreground">
                  Model: <span className="font-medium">{solveMutation.data.data.model}</span>
                </div>
              </div>
            ) : solveMutation.isPending ? (
              <div className="text-center py-12">
                <Loader2 className="h-12 w-12 mx-auto animate-spin text-primary" />
                <p className="mt-4 text-muted-foreground">{t('solving')}</p>
              </div>
            ) : (
              <div className="text-center py-12 text-muted-foreground">
                <ImageIcon className="h-12 w-12 mx-auto mb-4 opacity-50" />
                <p>Tải ảnh captcha lên để bắt đầu</p>
              </div>
            )}
          </div>

          {/* Usage Tips */}
          <div className="bg-card border rounded-lg p-4">
            <h3 className="font-medium mb-2">Mẹo sử dụng</h3>
            <ul className="text-sm text-muted-foreground space-y-1">
              <li>• Kéo thả hoặc paste ảnh trực tiếp</li>
              <li>• Hỗ trợ paste base64 từ clipboard</li>
              <li>• Chọn model phù hợp để tăng độ chính xác</li>
              <li>• Ảnh rõ nét sẽ cho kết quả tốt hơn</li>
            </ul>
          </div>
        </div>
      </div>
    </div>
  );
}