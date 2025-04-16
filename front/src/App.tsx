import { BrowserRouter as Router, Routes, Route } from 'react-router-dom';
import { ChatProvider } from '@/context';
import { MainLayout } from '@layouts';
import SearchPage from '@pages/Search';
import ResourcesPage from '@pages/Resources';
import '@/App.css'

function App() {
  return (
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
  );
}

export default App;
