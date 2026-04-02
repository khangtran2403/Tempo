
import React, { useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { useQuery, useMutation } from '@tanstack/react-query';
import toast from 'react-hot-toast';
import { ArrowLeft } from 'lucide-react';
import { workflowsAPI } from '../../api/workflows';
import WorkflowBuilder from '../../component/workflow/WorkflowBuilder';
import { WorkflowDefinition } from '../../types/workflow';

export default function WorkflowEditorPage() {
  const { id } = useParams<{ id?: string }>();
  const navigate = useNavigate();
  const [name, setName] = useState('');
  const [description, setDescription] = useState('');
  const [isActive, setIsActive] = useState(false);

  
  const { data: workflow, isLoading } = useQuery({
    queryKey: ['workflow', id],
    queryFn: () => (id ? workflowsAPI.getById(id) : null),
    enabled: !!id,
  });

 
  React.useEffect(() => {
    if (workflow) {
      setName(workflow.name);
      setDescription(workflow.description || '');
      setIsActive(workflow.is_active);
    }
  }, [workflow]);

  // Save mutation
  const saveMutation = useMutation({
    mutationFn: async (data: { name: string; description: string; definition: WorkflowDefinition; is_active: boolean }) => {
      if (id) {
        return workflowsAPI.update(id, data);
      } else {
        return workflowsAPI.create(data);
      }
    },
    onSuccess: (saved) => {
      toast.success(id ? 'Workflow được cập nhật!' : 'Workflow được tạo!');
      navigate(`/workflows/${saved.id}`);
    },
    onError: () => {
      toast.error('Lưu workflow thất bại. Vui lòng thử lại.');
    },
  });

  const handleSave = (definition: WorkflowDefinition) => {
    if (!name.trim()) {
      toast.error('Vui lòng nhập tên workflow.');
      return;
    }

    saveMutation.mutate({
      name,
      description,
      definition,
      is_active: isActive,
    });
  };

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-full">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary-600 mx-auto mb-4"></div>
          <p className="text-gray-600">Loading workflow...</p>
        </div>
      </div>
    );
  }

  return (
    <div className="h-full flex flex-col">
      {/* Header */}
      <div className="bg-white border-b border-gray-200 p-4">
        <div className="flex items-center justify-between mb-4">
          <button
            onClick={() => navigate('/workflows')}
            className="flex items-center space-x-2 text-gray-600 hover:text-gray-900"
          >
            <ArrowLeft className="w-5 h-5" />
            <span>Trở về</span>
          </button>
        </div>

        <div className="space-y-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Tên Workflow
            </label>
            <input
              type="text"
              value={name}
              onChange={(e) => setName(e.target.value)}
              className="input"
              placeholder="e.g., Send Daily Report"
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Mô tả Workflow 
            </label>
            <textarea
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              className="input"
              rows={2}
              placeholder="Mô tả chức năng của workflow..."
            />
          </div>

          <div className="flex items-center space-x-4">
            <label className="flex items-center space-x-2">
              <input
                type="checkbox"
                checked={isActive}
                onChange={(e) => setIsActive(e.target.checked)}
                className="w-4 h-4"
              />
              <span className="text-sm font-medium text-gray-700">
                Kích hoạt Workflow ngay sau khi lưu
              </span>
            </label>
          </div>
        </div>
      </div>

      {/* Builder */}
      <div className="flex-1 overflow-hidden">
        <WorkflowBuilder
          initialDefinition={workflow?.definition}
          onSave={handleSave}
        />
      </div>
    </div>
  );
}