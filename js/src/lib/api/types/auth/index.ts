export type User = {
  id: string;
  googleId: string;
  isAdmin: boolean;
  createdAt: string;
};

export type Session = {
  id: string;
  userId: string;
  createdAt: string;
  expiresAt: string;
};

export type AuthenticationObject = {
  user: User;
  session: Session;
};
