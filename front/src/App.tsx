import { BrowserRouter as Router, Routes, Route } from 'react-router-dom';
import { ChatProvider } from '@/context';
import { MainLayout } from '@layouts';
import SearchPage from '@pages/Search';
import ResourcesPage from '@pages/Resources';
import '@/App.css'
import {NotificationProvider} from "@/data/notification.context.tsx";

function App() {
  return (
      <NotificationProvider>
        <ChatProvider>
          <Router>
            <Routes>
              <Route path="/" element={<MainLayout />}>
                <Route index element={<SearchPage />} />
                <Route path="resources" element={<ResourcesPage />} />
              </Route>
            </Routes>
          </Router>
        </ChatProvider>
      </NotificationProvider>
  );
}

export default App;
