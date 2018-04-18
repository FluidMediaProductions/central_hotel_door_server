import React from 'react';
import PropTypes from 'prop-types';
import {Link} from "react-router-dom";
import {NavItem} from "reactstrap";

const MenuItem = ({ link, text }) => (
    <NavItem>
        <Link className="nav-link" to={link}>{text}</Link>
    </NavItem>
);

MenuItem.propTypes = {
    link: PropTypes.string.isRequired,
    text: PropTypes.string.isRequired
};

export default MenuItem
