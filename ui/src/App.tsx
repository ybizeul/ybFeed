import './App.css';
import { Button, Row, Col, Layout } from 'antd';
import { Content } from 'antd/es/layout/layout';

import {
  createBrowserRouter,
  RouterProvider,
} from "react-router-dom";

import React from 'react';

import Root from "./routes/root";
import Feed from "./routes/feed";

const router = createBrowserRouter([
  {
    path: "/",
    element: <Root />,
  },
  {
    path: "/:feed",
    element: <Feed />,
  },
]);

const App: React.FC = () => (
  <div className="App">
    <Layout>
        <Content>
        <Row>
          <Col span={6}/>
          <Col span={12}>
            <RouterProvider router={router} />
          </Col>
          <Col span={6}/>
        </Row>
      </Content>
    </Layout>
  </div>
);

export default App;