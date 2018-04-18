import React, {Component} from 'react';
import PropTypes from 'prop-types';
import {Button, Input} from 'reactstrap';
import makeGraphQLRequest from "../graphql";

class Login extends Component {
    constructor(props, context) {
        super(props, context);

        this.login = this.login.bind(this);
        this.state = {
            msg: ""
        }
    }

    login(e) {
        e.preventDefault();
        const login = this.refs.login.refs.value.value;
        const pass = this.refs.password.refs.value.value;

        const query = `
        mutation ($login: String!, $pass: String!) {
          loginUser(login: $login, pass: $pass)
        }`;
        const self = this;
        makeGraphQLRequest(query, {login: login, pass: pass}, function (resp) {
            if (resp["data"]["loginUser"] != null) {
                localStorage.setItem("JWTToken", resp["data"]["loginUser"]);
                if (typeof self.props.onLogin === "function") {
                    self.context.router.history.push('/');
                    self.props.onLogin();
                }
            } else {
                self.setState({
                    msg: "Invalid login"
                })
            }
        })
    }

    render() {
        return (
            <div className="h-100 text-center justify-content-center py-5 d-flex align-items-center">
                <form className="w-100" style={{maxWidth: 330}}>
                    <h1 className="h2 mb-3 font-weight-normal">Please sign in</h1>
                    <h4 className="text-danger">{this.state.msg}</h4>
                    <Input type="text" ref="login" placeholder="Username" className="rounded-0 border-bottom-0" innerRef="value"/>
                    <Input type="password" ref="password" placeholder="Password" className="rounded-0 mb-3" innerRef="value"/>
                    <Button size="lg" color="primary" onClick={this.login}>Sign in</Button>
                </form>
                <style>{
                    `html, body, #root {
                        height: 100%;
                    }
                    body {
                        background: #f5f5f5;
                    }
                `}</style>
            </div>
        );
    }
}

Login.propTypes = {
    onLogin: PropTypes.func
};

Login.contextTypes = {
    router: PropTypes.object
};

export default Login;