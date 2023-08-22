import './App.css';
import { Row, Col, Layout, ConfigProvider, theme } from 'antd';
import { useEffect, useState, useCallback } from "react";
import { Content } from 'antd/es/layout/layout';

import {
  createBrowserRouter,
  RouterProvider,
} from "react-router-dom";

import React from 'react';

import Root from "./routes/root";
import FeedComponent from "./routes/feed";

const router = createBrowserRouter([
  {
    path: "/",
    element: <Root />,
  },
  {
    path: "/:feed",
    element: <FeedComponent />,
  },
]);

const App: React.FC = () => {
  const [darkMode, setDarkMode] = useState(false);
  const windowQuery = window.matchMedia("(prefers-color-scheme:dark)");

  const darkModeChange = useCallback((event: MediaQueryListEvent) => {
    setDarkMode(event.matches ? true : false);
  }, []);

  useEffect(() => {
    windowQuery.addEventListener("change", darkModeChange);
    return () => {
      windowQuery.removeEventListener("change", darkModeChange);
    };
  }, [windowQuery, darkModeChange]);

  useEffect(() => {
    setDarkMode(windowQuery.matches ? true : false);
    // eslint-disable-line react-hooks/exhaustive-deps
  }, [windowQuery.matches]);

  return (
    <ConfigProvider
    theme={{
      algorithm: darkMode ? theme.darkAlgorithm : theme.compactAlgorithm
    }}
  >
  <div className="App">
    <Layout style={{minHeight:"100vh"}}>
        <Content>
        <Row>
          <Col xs={1} lg={6}/>
          <Col xs={22} lg={12}>
            <RouterProvider router={router} />
          </Col>
          <Col xs={1} lg={6}/>
        </Row>
      </Content>
    </Layout>
  </div>
  </ConfigProvider>
)};

export default App;