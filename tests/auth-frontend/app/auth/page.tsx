"use client";

import React, { useState, useEffect } from 'react';
import styles from './AuthPage.module.css';

// Definiciones de tipos para las respuestas de la API
interface UserInfo {
  id: string;
  name: string;
  email: string;
  role: string;
}

interface AuthStatusResponse {
  user: UserInfo;
  isAuthenticated: boolean;
}

interface MessageResponse {
  message: string;
}

interface ErrorResponse {
  error: string;
}

const API_BASE = 'http://localhost:8080/api/v1';

const AuthPage = () => {
  // Estado para los formularios
  const [registerEmail, setRegisterEmail] = useState('');
  const [registerName, setRegisterName] = useState('');
  const [registerPassword, setRegisterPassword] = useState('');
  const [registerRole, setRegisterRole] = useState('ESTUDIANTE');

  const [loginEmail, setLoginEmail] = useState('');
  const [loginPassword, setLoginPassword] = useState('');

  const [linkEmail, setLinkEmail] = useState('');
  const [linkPassword, setLinkPassword] = useState('');
  const [googleAuthCode, setGoogleAuthCode] = useState('');

  // Estado para los resultados
  const [registerResult, setRegisterResult] = useState<{ message: string; isError: boolean } | null>(null);
  const [loginResult, setLoginResult] = useState<{ message: string; isError: boolean } | null>(null);
  const [authStatusResult, setAuthStatusResult] = useState<{ message: string; isError: boolean } | null>(null);
  const [userInfo, setUserInfo] = useState<UserInfo | null>(null);
  const [isAuthenticated, setIsAuthenticated] = useState<boolean>(false);
  const [linkGoogleResult, setLinkGoogleResult] = useState<{ message: string; isError: boolean } | null>(null);

  // Configuración global para incluir cookies en todas las peticiones
  const fetchConfig: RequestInit = {
    credentials: 'include',
  };

  const showResult = (
    setter: React.Dispatch<React.SetStateAction<{ message: string; isError: boolean } | null>>,
    message: string,
    isError = false
  ) => {
    setter({ message, isError });
    // Limpiar el mensaje después de un tiempo si no es un error
    if (!isError) {
      setTimeout(() => setter(null), 5000);
    }
  };

  const handleRegisterSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setRegisterResult(null);

    const data = {
      email: registerEmail,
      name: registerName,
      password: registerPassword,
      role: registerRole,
    };

    try {
      const response = await fetch(`${API_BASE}/register`, {
        ...fetchConfig,
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(data),
      });
      const result: MessageResponse | ErrorResponse = await response.json();

      if (response.ok) {
        showResult(setRegisterResult, '¡Registro exitoso! Ahora puedes iniciar sesión.');
        setRegisterEmail('');
        setRegisterName('');
        setRegisterPassword('');
        setRegisterRole('ESTUDIANTE');
      } else {
        showResult(setRegisterResult, `Error: ${(result as ErrorResponse).error}`, true);
      }
    } catch (error) {
      showResult(setRegisterResult, 'Error de conexión', true);
      console.error('Error de registro:', error);
    }
  };

  const handleLoginSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoginResult(null);
    setUserInfo(null);
    setIsAuthenticated(false);

    const data = {
      email: loginEmail,
      password: loginPassword,
    };

    try {
      const response = await fetch(`${API_BASE}/login`, {
        ...fetchConfig,
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(data),
      });
      const result: MessageResponse | ErrorResponse = await response.json();

      if (response.ok) {
        showResult(setLoginResult, '¡Inicio de sesión exitoso! El token JWT ha sido establecido en una cookie.');
        setLoginEmail('');
        setLoginPassword('');
        // Vuelve a verificar el estado de autenticación después del login
        checkAuthStatus();
      } else {
        showResult(setLoginResult, `Error: ${(result as ErrorResponse).error}`, true);
      }
    } catch (error) {
      showResult(setLoginResult, 'Error de conexión', true);
      console.error('Error de login:', error);
    }
  };

  const checkAuthStatus = async () => {
    setAuthStatusResult(null);
    setUserInfo(null);
    setIsAuthenticated(false);

    try {
      const response = await fetch(`${API_BASE}/auth-status`, {
        ...fetchConfig,
        method: 'GET',
      });
      const result: AuthStatusResponse | ErrorResponse = await response.json();

      if (response.ok) {
        const authResponse = result as AuthStatusResponse;
        showResult(setAuthStatusResult, 'Estado de autenticación obtenido exitosamente');
        setUserInfo(authResponse.user);
        setIsAuthenticated(authResponse.isAuthenticated);
      } else {
        showResult(setAuthStatusResult, `Error: ${(result as ErrorResponse).error}`, true);
        setUserInfo(null);
        setIsAuthenticated(false);
      }
    } catch (error) {
      showResult(setAuthStatusResult, 'Error de conexión', true);
      setUserInfo(null);
      setIsAuthenticated(false);
      console.error('Error al obtener estado de autenticación:', error);
    }
  };

  const handleGoogleLogin = () => {
    window.location.href = `${API_BASE}/auth/google/login`;
  };

  const handleLogout = async () => {
    setLoginResult(null);
    setAuthStatusResult(null);
    setUserInfo(null);
    setIsAuthenticated(false);
    setLinkGoogleResult(null);

    try {
      const response = await fetch(`${API_BASE}/logout`, {
        ...fetchConfig,
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
      });
      const result: MessageResponse | ErrorResponse = await response.json();

      if (response.ok) {
        showResult(setAuthStatusResult, 'Sesión cerrada exitosamente.', false);
      } else {
        showResult(setAuthStatusResult, `Error al cerrar sesión: ${(result as ErrorResponse).error}`, true);
      }
    } catch (error) {
      showResult(setAuthStatusResult, 'Error de conexión al cerrar sesión.', true);
      console.error('Error al cerrar sesión:', error);
    }
  };

  const handleLinkGoogleAccount = async (e: React.FormEvent) => {
    e.preventDefault();
    setLinkGoogleResult(null);

    const data = {
      email: linkEmail,
      password: linkPassword,
      google_auth_code: googleAuthCode,
    };

    try {
      const response = await fetch(`${API_BASE}/auth/google/link`, {
        ...fetchConfig,
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(data),
      });
      const result: MessageResponse | ErrorResponse = await response.json();

      if (response.ok) {
        showResult(setLinkGoogleResult, 'Cuenta de Google vinculada exitosamente.');
        setLinkEmail('');
        setLinkPassword('');
        setGoogleAuthCode('');
      } else {
        showResult(setLinkGoogleResult, `Error: ${(result as ErrorResponse).error}`, true);
      }
    } catch (error) {
      showResult(setLinkGoogleResult, 'Error de conexión', true);
      console.error('Error al vincular cuenta de Google:', error);
    }
  };

  // Ejecutar checkAuthStatus al cargar la página
  useEffect(() => {
    checkAuthStatus();
  }, []);

  return (
    <div className={styles.body}>
      <h1 className={styles.h1}>Prueba de Autenticación</h1>

      <div className={styles.container}>
        {/* Registro */}
        <section id="register" className={styles.section}>
          <h2 className={styles.h2}>Registro</h2>
          <form onSubmit={handleRegisterSubmit}>
            <label className={styles.label}>
              Email:
              <input
                type="email"
                name="email"
                required
                className={styles.input}
                value={registerEmail}
                onChange={(e) => setRegisterEmail(e.target.value)}
              />
            </label>
            <label className={styles.label}>
              Nombre:
              <input
                type="text"
                name="name"
                required
                className={styles.input}
                value={registerName}
                onChange={(e) => setRegisterName(e.target.value)}
              />
            </label>
            <label className={styles.label}>
              Contraseña:
              <input
                type="password"
                name="password"
                required
                className={styles.input}
                value={registerPassword}
                onChange={(e) => setRegisterPassword(e.target.value)}
              />
            </label>
            <label className={styles.label}>
              Rol:
              <select
                name="role"
                required
                className={styles.select}
                value={registerRole}
                onChange={(e) => setRegisterRole(e.target.value)}
              >
                <option value="ESTUDIANTE">Estudiante</option>
                <option value="PROFESOR">Profesor</option>
              </select>
            </label>
            <button type="submit" className={styles.button}>
              Registrar
            </button>
          </form>
          {registerResult && (
            <div className={`${styles.result} ${registerResult.isError ? styles.error : styles.success}`}>
              {registerResult.message}
            </div>
          )}
        </section>

        {/* Login */}
        <section id="login" className={styles.section}>
          <h2 className={styles.h2}>Login</h2>
          <form onSubmit={handleLoginSubmit}>
            <label className={styles.label}>
              Email:
              <input
                type="email"
                name="email"
                required
                className={styles.input}
                value={loginEmail}
                onChange={(e) => setLoginEmail(e.target.value)}
              />
            </label>
            <label className={styles.label}>
              Contraseña:
              <input
                type="password"
                name="password"
                required
                className={styles.input}
                value={loginPassword}
                onChange={(e) => setLoginPassword(e.target.value)}
              />
            </label>
            <button type="submit" className={styles.button}>
              Iniciar Sesión
            </button>
          </form>
          {loginResult && (
            <div className={`${styles.result} ${loginResult.isError ? styles.error : styles.success}`}>
              {loginResult.message}
            </div>
          )}
        </section>
      </div>

      {/* Estado de Autenticación */}
      <section id="authStatus" className={styles.section}>
        <h2 className={styles.h2}>Estado de Autenticación</h2>
        <button onClick={checkAuthStatus} className={styles.button}>
          Obtener Estado de Autenticación
        </button>
        {authStatusResult && (
          <div className={`${styles.result} ${authStatusResult.isError ? styles.error : styles.success}`} style={{ color: '#000000' }}>
            {authStatusResult.message}
          </div>
        )}
        {userInfo && (
          <div className={styles['user-info']} style={{ display: userInfo ? 'block' : 'none', color: '#000000' }}>
            <h3>Información del Usuario:</h3>
            <pre style={{ color: '#000000' }}>{JSON.stringify({ user: userInfo, isAuthenticated }, null, 2)}</pre>
          </div>
        )}
        {isAuthenticated && (
          <button onClick={handleLogout} className={styles.button} style={{ marginTop: '1rem', backgroundColor: '#dc3545' }}>
            Cerrar Sesión
          </button>
        )}
      </section>

      {/* Google OAuth */}
      <section id="google" className={styles.section}>
        <h2 className={styles.h2}>Google OAuth</h2>
        <button onClick={handleGoogleLogin} className={styles.button}>
          Iniciar sesión con Google
        </button>
      </section>

      {/* Vinculación de cuenta Google */}
      <section id="linkGoogle" className={styles.section}>
        <h2 className={styles.h2}>Vincular Cuenta de Google</h2>
        <form onSubmit={handleLinkGoogleAccount}>
          <label className={styles.label}>
            Email:
            <input
              type="email"
              name="email"
              required
              className={styles.input}
              value={linkEmail}
              onChange={(e) => setLinkEmail(e.target.value)}
            />
          </label>
          <label className={styles.label}>
            Password:
            <input
              type="password"
              name="password"
              required
              className={styles.input}
              value={linkPassword}
              onChange={(e) => setLinkPassword(e.target.value)}
            />
          </label>
          <label className={styles.label}>
            Google Auth Code:
            <input
              type="text"
              name="google_auth_code"
              required
              className={styles.input}
              value={googleAuthCode}
              onChange={(e) => setGoogleAuthCode(e.target.value)}
            />
          </label>
          <button type="submit" className={styles.button}>
            Vincular Cuenta
          </button>
        </form>
        {linkGoogleResult && (
          <div className={`${styles.result} ${linkGoogleResult.isError ? styles.error : styles.success}`}>
            {linkGoogleResult.message}
          </div>
        )}
      </section>
    </div>
  );
};

export default AuthPage; 