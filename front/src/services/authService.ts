import Keycloak from 'keycloak-js';

class AuthService {
  private keycloak: Keycloak | null = null;
  private initialized = false;
  private initPromise: Promise<boolean> | null = null;

  constructor() {
    this.keycloak = new Keycloak({
      url: 'http://localhost:8080',
      realm: 'deltanotes',
      clientId: 'deltanotes-frontend',
    });
  }

  init(): Promise<boolean> {
    if (this.initialized) {
      return Promise.resolve(true);
    }

    if (this.initPromise) {
      return this.initPromise;
    }

    this.initPromise = new Promise((resolve, reject) => {
      if (!this.keycloak) {
        reject(new Error('Keycloak not initialized'));
        return;
      }

      this.keycloak
        .init({
          onLoad: 'check-sso',
          silentCheckSsoRedirectUri: window.location.origin + '/silent-check-sso.html',
          pkceMethod: 'S256',
        })
        .then((authenticated) => {
          this.initialized = true;
          resolve(authenticated);
        })
        .catch((error) => {
          console.error('Failed to initialize Keycloak', error);
          reject(error);
        });
    });

    return this.initPromise;
  }

  login(): void {
    if (this.keycloak) {
      this.keycloak.login();
    }
  }

  logout(): void {
    if (this.keycloak) {
      this.keycloak.logout();
    }
  }

  register(): void {
    if (this.keycloak) {
      this.keycloak.register();
    }
  }

  isAuthenticated(): boolean {
    return !!this.keycloak?.authenticated;
  }

  getToken(): string | undefined {
    return this.keycloak?.token;
  }

  getUsername(): string | undefined {
    return this.keycloak?.tokenParsed?.preferred_username;
  }

  updateToken(minValidity: number = 5): Promise<boolean> {
    if (!this.keycloak) {
      return Promise.reject(new Error('Keycloak not initialized'));
    }

    return this.keycloak.updateToken(minValidity);
  }
}

const authService = new AuthService();
export default authService; 