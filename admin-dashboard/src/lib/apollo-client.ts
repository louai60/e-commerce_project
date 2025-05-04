import { ApolloClient, InMemoryCache, HttpLink, from } from '@apollo/client';
import { onError } from '@apollo/client/link/error';

// Error handling link
const errorLink = onError(({ graphQLErrors, networkError }) => {
  if (graphQLErrors) {
    graphQLErrors.forEach(({ message, locations, path }) => {
      console.error(
        `[GraphQL error]: Message: ${message}, Location: ${locations}, Path: ${path}`
      );
    });
  }
  if (networkError) {
    console.error(`[Network error]: ${networkError}`);
  }
});

// HTTP link with authentication
const httpLink = new HttpLink({
  uri: `${process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api/v1'}/graphql`,
  credentials: 'include',
  headers: {
    'Content-Type': 'application/json',
  },
});

// Add auth token to requests
const authLink = new HttpLink({
  uri: `${process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api/v1'}/graphql`,
  credentials: 'include',
  headers: {
    'Content-Type': 'application/json',
  },
  fetchOptions: {
    mode: 'cors',
  },
});

// Create Apollo Client
export const apolloClient = new ApolloClient({
  link: from([errorLink, authLink]),
  cache: new InMemoryCache(),
  defaultOptions: {
    watchQuery: {
      fetchPolicy: 'cache-and-network',
      errorPolicy: 'all',
    },
    query: {
      fetchPolicy: 'network-only',
      errorPolicy: 'all',
    },
    mutate: {
      errorPolicy: 'all',
    },
  },
});
