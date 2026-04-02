import React, { useEffect } from 'react';
import { useNavigate, useSearchParams } from 'react-router-dom';
import { useAuthStore } from '../../store/authStore';
import toast from 'react-hot-toast';
import { User } from '../../types/user'; // Add this import

export default function GoogleAuthCallbackPage() {
  const navigate = useNavigate();
  const setAuth = useAuthStore((state) => state.setAuth);
  const [searchParams] = useSearchParams();

  useEffect(() => {
    const token = searchParams.get('token');
    const userStr = searchParams.get('user');
    const error = searchParams.get('error');

    if (error) {
      toast.error(error || 'Đăng nhập bằng Google thất bại. Vui lòng thử lại.');
      navigate('/login');
      return;
    }

    if (token && userStr) {
      try {
        const user: User = JSON.parse(decodeURIComponent(userStr));
        setAuth(token, user);
        toast.success('Đăng nhập bằng Google thành công!');
        navigate('/dashboard');
      } catch (e) {
        toast.error('Lỗi xử lý dữ liệu người dùng.');
        navigate('/login');
      }
    } else {
      toast.error('Không nhận được thông tin đăng nhập từ Google.');
      navigate('/login'); // Redirect to login page on error
    }
  }, [searchParams, setAuth, navigate]);

  return (
    <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-primary-50 to-primary-100">
      <div className="card w-full max-w-md text-center">
        <h1 className="text-2xl font-semibold text-primary-600 mb-4">Đang xử lý đăng nhập Google...</h1>
        <p className="text-gray-600">Vui lòng chờ giây lát.</p>
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary-600 mx-auto mt-6"></div>
      </div>
    </div>
  );
}
