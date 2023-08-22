import './App.css';
import { Row, Col, Layout, ConfigProvider, Space, theme } from 'antd';
import { useEffect, useState, useCallback } from "react";
import { Content, Footer } from 'antd/es/layout/layout';

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
  const [version, setVersion] = useState("");
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

  useEffect(() => {
    fetch("/api")
    .then(r => {
      let v = r.headers.get("Ybfeed-Version")
      if (v !== null) {
        setVersion(v)
      }
    })
  })
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
      <Footer>
        <Space style={{fontSize: '0.8em', opacity: '0.4', width: '100%', justifyContent: 'center'}}>
          ybFeed {version}
        </Space>
      </Footer>
    </Layout>
  </div>
  </ConfigProvider>
)};

export default App;