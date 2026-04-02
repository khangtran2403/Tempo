import React from 'react';
import { GitBranch } from 'lucide-react';
import toast from 'react-hot-toast';
const backendBaseUrl = (process.env.REACT_APP_API_URL || 'http://localhost:8080/api/v1').replace('/api/v1', '');

const services = [
  {
    name: 'Google',
    description: 'Connect Google Drive & Sheets.',
    icon: () => <img src="/assets/google-icon.svg" alt="Google" className="w-8 h-8" />,
    connectUrl: `${backendBaseUrl}/integrations/google/connect`,
  },
  {
    name: 'Notion',
    description: 'Connect your Notion workspace.',
    icon: () => <img src="/assets/notion-icon.svg" alt="Notion" className="w-8 h-8" />,
    connectUrl: `${backendBaseUrl}/integrations/notion/connect`,
  },
  {
    name: 'GitHub',
    description: 'Connect your GitHub account.',
    icon: GitBranch,
    connectUrl: `${backendBaseUrl}/integrations/github/connect`,
  },
];


export default function AddIntegrationPage() {
  
  const handleConnect = (url: string) => {
    const token = localStorage.getItem('token');
    if (!token) {
      toast.error("You are not logged in. Please log in again.");
      return;
    }
    
    // Append token to the URL and redirect
    const connectUrlWithToken = `${url}?token=${token}`;
    window.location.href = connectUrlWithToken;
  };

  return (
    <div className="p-6 space-y-6">
      <div className="max-w-3xl mx-auto">
        <h1 className="text-3xl font-bold text-gray-900">Thêm tích hợp mới</h1>
        <p className="text-gray-600 mt-1">Chọn một dịch vụ để kết nối với Tempo.</p>
        
        <div className="mt-8 space-y-4">
          {services.map((service) => (
            <div key={service.name} className="card flex items-center justify-between p-4">
              <div className="flex items-center space-x-4">
                <service.icon className="w-8 h-8" />
                <div>
                  <h3 className="font-semibold text-lg">{service.name}</h3>
                  <p className="text-sm text-gray-500">{service.description}</p>
                </div>
              </div>
              <button onClick={() => handleConnect(service.connectUrl)} className="btn btn-primary">
                Kết nối
              </button>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}
