import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { BrowserRouter } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import toast from 'react-hot-toast';
import LoginPage from './LoginPage';
import { authAPI } from '../../api/auth';
import { useAuthStore } from '../../store/authStore';

// Mock dependencies
jest.mock('../../api/auth');
jest.mock('react-hot-toast');

// Mock the part of the store we need and the navigation
const mockedAuthAPI = authAPI as jest.Mocked<typeof authAPI>;
const mockedToast = toast as jest.Mocked<typeof toast>;
const mockedNavigate = jest.fn();

jest.mock('react-router-dom', () => ({
  ...jest.requireActual('react-router-dom'),
  useNavigate: () => mockedNavigate,
}));

const queryClient = new QueryClient();

const renderComponent = () => {
  return render(
    <QueryClientProvider client={queryClient}>
      <BrowserRouter>
        <LoginPage />
      </BrowserRouter>
    </QueryClientProvider>
  );
};

describe('LoginPage', () => {
  beforeEach(() => {
    // Reset mocks and store before each test
    jest.clearAllMocks();
    useAuthStore.setState({
      user: null,
      token: null,
      isAuthenticated: false,
    });
  });

  it('renders the login form correctly', () => {
    renderComponent();
    expect(screen.getByLabelText(/email/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/mật khẩu/i)).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /đăng nhập/i })).toBeInTheDocument();
  });

  it('shows validation error for invalid email', async () => {
    renderComponent();
    fireEvent.change(screen.getByLabelText(/email/i), { target: { value: 'invalid-email' } });
    fireEvent.click(screen.getByRole('button', { name: /đăng nhập/i }));

    await waitFor(() => {
      expect(mockedToast.error).toHaveBeenCalledWith('Email không hợp lệ ');
    });
    expect(mockedAuthAPI.login).not.toHaveBeenCalled();
  });
  
    it('shows validation error for empty password', async () => {
    renderComponent();
    fireEvent.change(screen.getByLabelText(/email/i), { target: { value: 'test@example.com' } });
    fireEvent.click(screen.getByRole('button', { name: /đăng nhập/i }));

    await waitFor(() => {
      expect(mockedToast.error).toHaveBeenCalledWith('Vui lòng nhập mật khẩu');
    });
    expect(mockedAuthAPI.login).not.toHaveBeenCalled();
  });

  it('handles successful login', async () => {
    const mockUser = { id: '1', name: 'Test User', email: 'test@example.com', created_at: '' };
    mockedAuthAPI.login.mockResolvedValue({ token: 'fake-token', user: mockUser });
    const setAuth = jest.spyOn(useAuthStore.getState(), 'setAuth');

    renderComponent();

    fireEvent.change(screen.getByLabelText(/email/i), { target: { value: 'test@example.com' } });
    fireEvent.change(screen.getByLabelText(/mật khẩu/i), { target: { value: 'password123' } });
    fireEvent.click(screen.getByRole('button', { name: /đăng nhập/i }));

    await waitFor(() => {
      expect(mockedAuthAPI.login).toHaveBeenCalledWith('test@example.com', 'password123');
      expect(setAuth).toHaveBeenCalledWith('fake-token', mockUser);
      expect(mockedToast.success).toHaveBeenCalledWith('Đăng nhập thành công!');
      expect(mockedNavigate).toHaveBeenCalledWith('/dashboard');
    });
  });

  it('handles failed login', async () => {
    const errorMessage = 'Invalid credentials';
    mockedAuthAPI.login.mockRejectedValue({ message: errorMessage });
    const setAuth = jest.spyOn(useAuthStore.getState(), 'setAuth');
    
    renderComponent();

    fireEvent.change(screen.getByLabelText(/email/i), { target: { value: 'test@example.com' } });
    fireEvent.change(screen.getByLabelText(/mật khẩu/i), { target: { value: 'wrongpassword' } });
    fireEvent.click(screen.getByRole('button', { name: /đăng nhập/i }));

    await waitFor(() => {
      expect(mockedAuthAPI.login).toHaveBeenCalledWith('test@example.com', 'wrongpassword');
      expect(setAuth).not.toHaveBeenCalled();
      expect(mockedToast.error).toHaveBeenCalledWith(errorMessage);
      expect(mockedNavigate).not.toHaveBeenCalled();
    });
  });
});
