import makeGraphQLRequest from "./graphql";
import * as jwt from "jsonwebtoken";

export function getJWT() {
    const token = localStorage.getItem("JWTToken");
    const decoded = jwt.decode(token);
    if (decoded != null) {
        const expiry = decoded.exp;
        const now = Math.round((new Date()).getTime() / 1000);
        if ((expiry - now) < 86400) {
            const query = `
        mutation ($token: String!) {
          refreshToken(token: $token)
        }`;
            makeGraphQLRequest(query, {token: token}, function (resp) {
                if (resp["data"]["refreshToken"] != null) {
                    localStorage.setItem("JWTToken", resp["data"]["refreshToken"])
                }
            })
        }
        return token
    } else {
        return ""
    }
}

export function delJWT() {
    localStorage.removeItem("JWTToken")
}

export function haveJWT() {
    return getJWT() != null
}

export function JWTIsValid(callback) {
    if (haveJWT()) {
        const query = `
        query ($token: String!) {
          auth(token: $token) {
            self {
              name
            }
          }
        }`;
        makeGraphQLRequest(query, {token: getJWT()}, function (resp) {
            if (resp["data"]["auth"] != null) {
                callback(true);
            } else {
                callback(false);
            }
        })
    } else {
        callback(false)
    }
}