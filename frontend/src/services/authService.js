const API_BASE_URL = '/api/v1/auth';

async function handleResponse(response) {
  if (!response.ok) {
    const message = await response.text();
    throw new Error(message || 'Something went wrong');
  }

  return response.json();
}

export async function registerUser(userData) {
  const response = await fetch(`${API_BASE_URL}/register`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(userData),
  });

  return handleResponse(response);
}

export async function loginUser(credentials) {
  const response = await fetch(`${API_BASE_URL}/login`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(credentials),
  });

  const data = await handleResponse(response);

  localStorage.setItem('access_token', data.access_token);

  return data;
}

export function logoutUser() {
  localStorage.removeItem('access_token');
}

export function getAccessToken() {
  return localStorage.getItem('access_token');
}
