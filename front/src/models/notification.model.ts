export interface Notification {
  id: string;
  message: string;
  type: 'error' | 'warning' | 'info' | 'success';
  timeout?: number;
}

export interface NotificationContextType {
  notifications: Notification[];
  addNotification: (message: string, type: Notification['type'], timeout?: number) => void;
  removeNotification: (id: string) => void;
} 