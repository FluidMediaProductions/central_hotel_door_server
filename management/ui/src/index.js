import React from 'react';
import ReactDOM from 'react-dom';
import { BrowserRouter, withRouter } from 'react-router-dom'
import App from './App';
import registerServiceWorker from './registerServiceWorker';
import './index.css';
import 'bootstrap/dist/css/bootstrap.min.css';

const RouterApp = withRouter(App);
ReactDOM.render(
    <BrowserRouter>
        <RouterApp/>
    </BrowserRouter>,
    document.getElementById('root'));
registerServiceWorker();
