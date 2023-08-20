import React from 'react';
import ReactDOM from 'react-dom/client';
import './index.css';
import App from './App';
import reportWebVitals from './reportWebVitals';
import { StyleProvider } from '@ant-design/cssinjs';

const root = ReactDOM.createRoot(
  document.getElementById('root') as HTMLElement
);

// var t = theme.defaultAlgorithm
// if (window.matchMedia && window.matchMedia('(prefers-color-scheme: dark)').matches) {
//   t = theme.darkAlgorithm
// }
root.render(
  <React.StrictMode>
    <StyleProvider hashPriority="high">
      <App />
    </StyleProvider>
  </React.StrictMode>
);

// If you want to start measuring performance in your app, pass a function
// to log results (for example: reportWebVitals(console.log))
// or send to an analytics endpoint. Learn more: https://bit.ly/CRA-vitals
reportWebVitals();
