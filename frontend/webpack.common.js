const path = require("path");
const HtmlWebpackPlugin = require("html-webpack-plugin");
const CopyPlugin = require("copy-webpack-plugin");
const MiniCssExtractPlugin = require("mini-css-extract-plugin");

module.exports = {
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
        new MiniCssExtractPlugin()
    ],    
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
                use: [
                    MiniCssExtractPlugin.loader,
                    {
                        loader: "css-loader", // translates CSS into CommonJS modules
                    },
                    {
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
