import './App.css';
import { Row, Col, Layout } from 'antd';
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
  <div className="App" style={{
    height:'100%'}}>
    <Layout>
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
);

export default App;