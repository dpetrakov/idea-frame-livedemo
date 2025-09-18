import React from 'react';
import { useAuth } from '../contexts/AuthContext';
import './Dashboard.css';

export function Dashboard() {
  const { user, logout } = useAuth();

  return (
    <div className="dashboard-container">
      <header className="dashboard-header">
        <h1>Idea Frame</h1>
        <div className="user-info">
          <span>Привет, {user?.displayName}!</span>
          <button onClick={logout} className="logout-button">
            Выйти
          </button>
        </div>
      </header>
      
      <main className="dashboard-main">
        <div className="welcome-card">
          <h2>Добро пожаловать в систему управления инициативами!</h2>
          <p>
            Здесь будет реализован функционал создания и управления инициативами.
            Пока что вы успешно прошли аутентификацию.
          </p>
          
          <div className="user-details">
            <h3>Информация о пользователе:</h3>
            <ul>
              <li><strong>ID:</strong> {user?.id}</li>
              <li><strong>Логин:</strong> {user?.login}</li>
              <li><strong>Имя:</strong> {user?.displayName}</li>
              <li><strong>Регистрация:</strong> {user?.createdAt ? new Date(user.createdAt).toLocaleDateString('ru-RU') : 'Неизвестно'}</li>
            </ul>
          </div>
        </div>
      </main>
    </div>
  );
}