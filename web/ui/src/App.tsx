// Import styles of packages that you've installed.
// All packages except `@mantine/hooks` require styles imports
import '@mantine/core/styles.css';
import '@mantine/notifications/styles.css';
import '@mantine/dropzone/styles.css';

import { MantineProvider, createTheme } from '@mantine/core';
import { Notifications } from '@mantine/notifications';

import { YBFeedApp } from './YBFeed/YBFeedApp'

const theme = createTheme({
  primaryColor: "gray",
});

export default function App() {
  return (
    <MantineProvider theme={theme} defaultColorScheme='auto'>
      <Notifications position='top-center' autoClose={2000}/>
      <YBFeedApp/>
    </MantineProvider>
  )
}