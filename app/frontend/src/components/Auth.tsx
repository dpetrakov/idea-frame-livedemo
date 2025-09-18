import React, { useState } from 'react';
import { useAuth } from '../contexts/AuthContext';
import { UserRegisterRequest, UserLoginRequest } from '../types/auth';
import './Auth.css';

export function Auth() {
  const [isLogin, setIsLogin] = useState(true);
  const [formData, setFormData] = useState({
    login: '',
    displayName: '',
    password: '',
    confirmPassword: '',
  });

  const { login, register, isLoading, error, clearError } = useAuth();

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const { name, value } = e.target;
    setFormData(prev => ({
      ...prev,
      [name]: value
    }));
    
    // Очищаем ошибку при изменении полей
    if (error) {
      clearError();
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    
    try {
      if (isLogin) {
        const loginData: UserLoginRequest = {
          login: formData.login,
          password: formData.password,
        };
        await login(loginData);
      } else {
        const registerData: UserRegisterRequest = {
          login: formData.login,
          displayName: formData.displayName,
          password: formData.password,
          confirmPassword: formData.confirmPassword,
        };
        await register(registerData);
      }
    } catch (error) {
      // Ошибка уже обработана в контексте
    }
  };

  const switchMode = () => {
    setIsLogin(!isLogin);
    setFormData({
      login: '',
      displayName: '',
      password: '',
      confirmPassword: '',
    });
    clearError();
  };

  const isFormValid = () => {
    if (isLogin) {
      return formData.login.length >= 3 && formData.password.length >= 8;
    } else {
      return (
        formData.login.length >= 3 &&
        formData.displayName.length >= 1 &&
        formData.password.length >= 8 &&
        formData.password === formData.confirmPassword
      );
    }
  };

  return (
    <div className="auth-container">
      <div className="auth-card">
        <h1 className="auth-title">Idea Frame</h1>
        
        <div className="auth-tabs">
          <button 
            className={`tab ${isLogin ? 'active' : ''}`}
            onClick={() => isLogin || switchMode()}
          >
            Вход
          </button>
          <button 
            className={`tab ${!isLogin ? 'active' : ''}`}
            onClick={() => !isLogin || switchMode()}
          >
            Регистрация
          </button>
        </div>

        <form onSubmit={handleSubmit} className="auth-form">
          <div className="form-group">
            <label htmlFor="login">Логин</label>
            <input
              id="login"
              type="text"
              name="login"
              value={formData.login}
              onChange={handleInputChange}
              placeholder="Введите логин (3-32 символа)"
              required
              minLength={3}
              maxLength={32}
              pattern="[a-zA-Z0-9_-]+"
              disabled={isLoading}
            />
          </div>

          {!isLogin && (
            <div className="form-group">
              <label htmlFor="displayName">Отображаемое имя</label>
              <input
                id="displayName"
                type="text"
                name="displayName"
                value={formData.displayName}
                onChange={handleInputChange}
                placeholder="Введите имя (1-32 символа)"
                required
                minLength={1}
                maxLength={32}
                disabled={isLoading}
              />
            </div>
          )}

          <div className="form-group">
            <label htmlFor="password">Пароль</label>
            <input
              id="password"
              type="password"
              name="password"
              value={formData.password}
              onChange={handleInputChange}
              placeholder="Введите пароль (8-64 символа)"
              required
              minLength={8}
              maxLength={64}
              disabled={isLoading}
            />
          </div>

          {!isLogin && (
            <div className="form-group">
              <label htmlFor="confirmPassword">Подтверждение пароля</label>
              <input
                id="confirmPassword"
                type="password"
                name="confirmPassword"
                value={formData.confirmPassword}
                onChange={handleInputChange}
                placeholder="Повторите пароль"
                required
                minLength={8}
                maxLength={64}
                disabled={isLoading}
              />
              {formData.confirmPassword && formData.password !== formData.confirmPassword && (
                <div className="field-error">Пароли не совпадают</div>
              )}
            </div>
          )}

          {error && (
            <div className="error-message">
              {error}
            </div>
          )}

          <button 
            type="submit" 
            className="submit-button"
            disabled={isLoading || !isFormValid()}
          >
            {isLoading ? 'Загрузка...' : (isLogin ? 'Войти' : 'Зарегистрироваться')}
          </button>
        </form>
      </div>
    </div>
  );
}