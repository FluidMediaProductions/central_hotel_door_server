import React from 'react';
import PropTypes from 'prop-types';
import {Pagination, PaginationItem} from "reactstrap";

const CustomPagination = ({previousDisabled, nextDisabled, previousPage, nextPage}) => (
    <Pagination className="justify-content-center">
        <PaginationItem disabled={previousDisabled}>
            <a href="" className="page-link" onClick={previousPage}>Previous</a>
        </PaginationItem>
        <PaginationItem disabled={nextDisabled}>
            <a href="" className="page-link" onClick={nextPage}>Next</a>
        </PaginationItem>
    </Pagination>
);

Pagination.propTypes = {
    previousDisabled: PropTypes.bool.isRequired,
    nextDisabled: PropTypes.bool.isRequired,
    previousPage: PropTypes.func.isRequired,
    nextPage: PropTypes.func.isRequired,
};

export default CustomPagination;