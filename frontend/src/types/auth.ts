export interface User {
  id: string;
  email: string;
  username: string;
  firstName: string;
  lastName: string;
  phoneNumber?: string;
  role: string;
  createdAt: string;
}

export interface LoginCredentials {
  email: string;
  password: string;
}

export interface RegisterCredentials {
  Email: string;
  Username: string;
  Password: string;
  FirstName: string;
  LastName: string;
  PhoneNumber?: string;
}

export interface AuthResponse {
  access_token: string;
  refresh_token: string;
  user: User;
}

