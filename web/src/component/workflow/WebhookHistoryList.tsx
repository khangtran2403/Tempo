// src/component/workflow/WebhookHistoryList.tsx
import React from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { RefreshCw, Play, AlertTriangle } from 'lucide-react';
import toast from 'react-hot-toast';
import { webhooksAPI } from '../../api/webhooks';
import { WebhookHistory } from '../../types/webhook';

interface WebhookHistoryListProps {
  workflowId: string;
}

export default function WebhookHistoryList({ workflowId }: WebhookHistoryListProps) {
  const queryClient = useQueryClient();

  const { data: historyData, isLoading, isError, error } = useQuery({
    queryKey: ['webhookHistory', workflowId],
    queryFn: () => webhooksAPI.listByWorkflowId(workflowId),
  });

  const replayMutation = useMutation({
    mutationFn: ({ workflowId, historyId }: { workflowId: string; historyId: string }) => 
      webhooksAPI.replay(workflowId, historyId),
    onSuccess: (data) => {
      toast.success(`Successfully replayed webhook. New execution ID: ${data.execution_id}`);
      queryClient.invalidateQueries({ queryKey: ['executions', workflowId] });
    },
    onError: (err: any) => {
      toast.error(`Failed to replay webhook: ${err.message || 'Unknown error'}`);
    },
  });

  if (isLoading) {
    return <div className="text-center py-8">Tải lịch sử gọi webhook...</div>;
  }

  if (isError) {
    return (
      <div className="text-center py-8 text-red-600">
        <AlertTriangle className="w-8 h-8 mx-auto mb-2" />
        <p>Lỗi khi tải dữ liệu: {error.message}</p>
      </div>
    );
  }

  const history = historyData?.history || [];

  return (
    <div className="space-y-3">
      {history.length === 0 ? (
        <div className="text-center py-12 text-gray-500">
          Chưa có lịch sử gọi webhook nào.
        </div>
      ) : (
        history.map((item) => (
          <WebhookHistoryItem
            key={item.primaryKey}
            item={item}
            onReplay={() => replayMutation.mutate({ workflowId, historyId: item.primaryKey })}
            isReplaying={replayMutation.isPending && replayMutation.variables?.historyId === item.primaryKey}
          />
        ))
      )}
    </div>
  );
}

interface WebhookHistoryItemProps {
  item: WebhookHistory;
  onReplay: () => void;
  isReplaying: boolean;
}

function WebhookHistoryItem({ item, onReplay, isReplaying }: WebhookHistoryItemProps) {
  // Defensive check for status
  const status = item.status || 'unknown';
  const statusColor = status === 'success' ? 'text-green-600' : 'text-red-600';
  const statusBg = status === 'success' ? 'bg-green-100' : 'bg-red-100';

  return (
    <div className="p-4 border border-gray-200 rounded-lg transition-all hover:border-primary-300 hover:bg-gray-50">
      <div className="flex items-center justify-between">
        <div className="flex-1 space-y-1">
          <div className="flex items-center space-x-3">
            <span className={`px-2 py-0.5 rounded text-xs font-semibold ${statusBg} ${statusColor}`}>
              {status.toUpperCase()}
            </span>
            <span className="text-sm font-medium text-gray-800">{item.method}</span>
            <span className="text-xs text-gray-500 font-mono">{item.primaryKey}</span>
          </div>
          <p className="text-sm text-gray-600">
            Nhận vào lúc: {new Date(item.received_at).toLocaleString()} từ {item.ip_address}
          </p>
          {status === 'failed' && item.error && (
            <p className="text-xs text-red-600">Lỗi: {item.error}</p>
          )}
        </div>
        <div className="flex items-center">
          <button
            onClick={onReplay}
            disabled={isReplaying}
            className="p-2 rounded-lg hover:bg-gray-200 disabled:opacity-50 disabled:cursor-wait"
            title="Replay Webhook"
          >
            {isReplaying ? (
              <RefreshCw className="w-5 h-5 text-primary-600 animate-spin" />
            ) : (
              <Play className="w-5 h-5 text-gray-600" />
            )}
          </button>
        </div>
      </div>
      {/* Optionally, add a section to expand and view headers/body */}
    </div>
  );
}
