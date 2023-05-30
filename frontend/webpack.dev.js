// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

const { merge } = require("webpack-merge");
const common = require("./webpack.common.js");

module.exports = merge(common, {
    mode: "development",
    devtool: "inline-source-map",
    devServer: {
        compress: true,
        port: 9000,
        static: {
            directory: path.join(__dirname, "dist"),
        },
        proxy: {
            "/api": "http://localhost:8080"
        }
    },
});
