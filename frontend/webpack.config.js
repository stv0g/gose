const path = require("path");
const HtmlWebpackPlugin = require("html-webpack-plugin");
const CopyPlugin = require("copy-webpack-plugin");

module.exports = {
    mode: "development",
    entry: "./src/index.ts",
    output: {
        filename: "main.js",
        path: path.resolve(__dirname, "dist"),
    },
    plugins: [
        new HtmlWebpackPlugin({
            title: "GoS3 - A terascale file uploader",
            template: "index.html",
        }),
        new CopyPlugin({
            patterns: [
              { from: "img/*", to: "" }
            ],
          }),
    ],
    devtool: "eval-source-map",
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
    resolve: {
        modules: ["node_modules"],
        extensions: [".tsx", ".ts", ".js"]
    },
    module: {
        rules: [
            {
                test: /\.tsx?$/,
                use: "ts-loader",
                exclude: /node_modules/,
            },
            {
                test: /\.(scss)$/,
                use: [{
                        loader: "style-loader", // inject CSS to page
                    }, {
                        loader: "css-loader", // translates CSS into CommonJS modules
                    }, {
                        loader: "postcss-loader", // Run post css actions
                        options: {
                            postcssOptions: {
                                plugins() { // post css plugins, can be exported to postcss.config.js
                                    return [
                                        require("precss"),
                                        require("autoprefixer")
                                    ];
                                }
                            }
                        },
                    },
                    {
                        loader: "sass-loader" // compiles Sass to CSS
                    }
                ]
            },
        ]
    },
};
