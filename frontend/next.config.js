/** @type {import('next').NextConfig} */
const nextConfig = {
  server: {
    https: {
      // You can specify custom certificate paths
      cert: './certificates/localhost.crt',
      key: './certificates/localhost.key',
    },
  },
};

module.exports = nextConfig;

  