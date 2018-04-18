import React, {Component} from 'react';
import PropTypes from 'prop-types';
import MenuItem from './MenuItem';
import {Collapse, Nav, Navbar, NavbarBrand, NavbarToggler} from "reactstrap";

class Menu extends Component {
    constructor(props) {
        super(props);

        this.toggle = this.toggle.bind(this);
        this.state = {
            isOpen: false
        };
    }

    toggle() {
        this.setState({
            isOpen: !this.state.isOpen
        });
    }

    render() {
        return (
            <Navbar expand="md" light>
                <NavbarBrand href="/">Hotel door system</NavbarBrand>
                <NavbarToggler onClick={this.toggle}/>
                <Collapse isOpen={this.state.isOpen} navbar>
                    <Nav navbar>
                        {this.props.pages.map(page => (
                            <MenuItem key={page.id} text={page.title} link={page.link}/>
                        ))}
                    </Nav>
                </Collapse>
            </Navbar>
        )
    }
}

Menu.propTypes = {
    pages: PropTypes.arrayOf(
        PropTypes.shape({
            id: PropTypes.number.isRequired,
            title: PropTypes.string.isRequired,
            link: PropTypes.string.isRequired
        }).isRequired
    ).isRequired
};

export default Menu;