const path = require('path');
const HtmlWebpackPlugin = require('html-webpack-plugin');


module.exports = {
    mode: 'development',
    entry: './src/index.js',
    output: {
        filename: 'main.js',
        path: path.resolve(__dirname, 'dist'),
    },
    plugins: [
        new HtmlWebpackPlugin({
            title: 'GoS3 - A terascale file uploader',
            template: 'index.html',
            favicon: 'img/gose-logo.svg'
        })
    ],
    devtool: 'eval-source-map', 
    devServer: {
        compress: true,
        port: 9000,
        static: {
            directory: path.join(__dirname, 'dist'),
        },
        proxy: {
            '/api': 'http://localhost:8080'
        }
    },
    resolve: {
        modules: ['node_modules'],
        fallback: {
            crypto: require.resolve("crypto-browserify"),
            stream: require.resolve("stream-browserify"),
            buffer: require.resolve("buffer/"),
        },
    },
    module: {
        rules: [
            //   {
            //     test: /\.m?js$/,
            //     exclude: /node_modules/,
            //     use: {
            //       loader: "babel-loader",
            //       options: {
            //         presets: ['@babel/preset-env']
            //       }
            //     }
            //   },
            {
                test: /\.html$/i,
                loader: "html-loader",
            },
            {
                test: /\.(scss)$/,
                use: [
                    {
                        loader: 'style-loader', // inject CSS to page
                    }, {
                        loader: 'css-loader', // translates CSS into CommonJS modules
                    }, {
                        loader: 'postcss-loader', // Run post css actions
                        options: {
                            postcssOptions: {
                                plugins: function () { // post css plugins, can be exported to postcss.config.js
                                    return [
                                        require('precss'),
                                        require('autoprefixer')
                                    ];
                                }
                            }
                        },
                    },
                    {
                        loader: 'sass-loader' // compiles Sass to CSS
                    }
                ]
            },
        ]
    },
};
