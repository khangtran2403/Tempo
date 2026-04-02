
import React, { useState } from 'react';
import { useNavigate, Link } from 'react-router-dom';
import { authAPI } from '../../api/auth';
import { useAuthStore } from '../../store/authStore';
import toast from 'react-hot-toast';

export default function RegisterPage() {
  const [name, setName] = useState('');
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [loading, setLoading] = useState(false);
  
  const navigate = useNavigate();
  const setAuth = useAuthStore((state) => state.setAuth);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!name.trim()) {
      toast.error('Vui lòng nhập tên');
      return;
    }
    if (!email || !/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email)) {
      toast.error('Email không hợp lệ');
      return;
    }
    if (password.length < 6) {
      toast.error('Mật khẩu phải có ít nhất 6 ký tự');
      return;
    }

    setLoading(true);

    try {
      const response = await authAPI.register(email, password, name);
      setAuth(response.token, response.user);
      toast.success('Tạo tài khoản thành công!');
      navigate('/dashboard');
    } catch (error: any) {
      toast.error(error?.message || 'Đăng ký thất bại');
    } finally {
      setLoading(false);
    }
  };

  const onGoogle = () => {
    const base = process.env.REACT_APP_API_URL || 'http://localhost:8080/api/v1';
    window.location.href = base.replace(/\/api\/v1$/, '') + '/auth/google/login';
  };

  return (
    <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-primary-50 to-primary-100">
      <div className="card w-full max-w-md">
        <div className="text-center mb-8">
          <h1 className="text-3xl font-bold text-primary-600">Tempo</h1>
          <p className="text-gray-600 mt-2">Đăng ký tài khoản</p>
        </div>

        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Tên
            </label>
            <input
              type="text"
              value={name}
              onChange={(e) => setName(e.target.value)}
              className="input"
              placeholder="John Doe"
              required
              disabled={loading}
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Email
            </label>
            <input
              type="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              className="input"
              placeholder="you@example.com"
              required
              disabled={loading}
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Mật khẩu
            </label>
            <input
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              className="input"
              placeholder="••••••••"
              minLength={6}
              required
              disabled={loading}
            />
            <p className="text-xs text-gray-500 mt-1">
              Ít nhất 6 ký tự
            </p>
          </div>

          <button
            type="submit"
            disabled={loading}
            className="btn btn-primary w-full"
          >
            {loading ? 'Đang tạo tài khoản...' : 'Đăng ký'}
          </button>
        </form>

        <div className="my-6 flex items-center">
          <div className="flex-1 h-px bg-gray-200" />
          <span className="px-3 text-xs text-gray-500">HOẶC</span>
          <div className="flex-1 h-px bg-gray-200" />
        </div>

        <button onClick={onGoogle} className="btn btn-secondary w-full" disabled={loading}>
          Đăng ký bằng Google
        </button>

        <p className="text-center text-sm text-gray-600 mt-6">
          Đã có tài khoản?{' '}
          <Link to="/login" className="text-primary-600 hover:text-primary-700 font-medium">
            Đăng nhập
          </Link>
        </p>
      </div>
    </div>
  );
}