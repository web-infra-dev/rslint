import React from 'react';
import { Routes, Route } from 'react-router-dom';
import { ThemeProvider } from './components/theme-provider';
import HomePage from './pages/HomePage';
import DevPage from './pages/DevPage';

function App() {
  return (
    <ThemeProvider
      attribute="class"
      defaultTheme="system"
      enableSystem
      disableTransitionOnChange
    >
      <div className="bg-background text-foreground rslint">
        <main id="root">
          <Routes>
            <Route path="/" element={<HomePage />} />
            <Route path="/dev" element={<DevPage />} />
          </Routes>
        </main>
      </div>
    </ThemeProvider>
  );
}

export default App;
