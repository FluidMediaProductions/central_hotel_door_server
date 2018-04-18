import PropTypes from "prop-types";

function makeGraphQLRequest(query, variables, callback) {
    const xhr = new XMLHttpRequest();
    xhr.responseType = 'json';
    xhr.open("POST", "http://localhost:3002/graphql");
    xhr.setRequestHeader("Content-Type", "application/json");
    xhr.setRequestHeader("Accept", "application/json");
    xhr.onload = function () {
        if (xhr.responseType === "json") {
            if (xhr.response.errors != null) {
                for (let i in xhr.response.errors) {
                    const error = xhr.response.errors[i];
                    for (let t in error.locations) {
                        const location = error.locations[t];
                        console.error("GraphQL error at " + location.line + ":" + location.column + ": " + error.message);
                    }
                }
            }
            callback(xhr.response)
        }
    };
    xhr.send(JSON.stringify({
        query: query,
        variables: variables
    }));
}

makeGraphQLRequest.propTypes = {
    query: PropTypes.string.isRequired,
    variables: PropTypes.object.isRequired,
    callback: PropTypes.func.isRequired
};

export default makeGraphQLRequest;