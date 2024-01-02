import { Navigate } from 'react-router';
import { useState } from "react"
import { Text, TextInput, Container } from '@mantine/core';
import { useForm } from '@mantine/form';
import { YBBreadCrumbComponent } from "./Components/YBBreadCrumbComponent";


export function YBFeedHome() {
    const [goToPath,setGoToPath] = useState("")
    
    const form = useForm({
        initialValues: {
          feedName: '',
        },
    
        validate: {
            feedName: (value) => (/^[a-zA-Z0-9]+$/.test(value) ? null : 'Invalid feed name'),
        },
      });

    const handleFinish = (feedName:string) => {
        setGoToPath("/" + encodeURIComponent(feedName))
    }

    // if (window.location.pathname !== "/") {
    //     setGoToPath(window.location.pathname)
    // }
  
    return (
        <>
        {goToPath && (
            <Navigate to={goToPath} replace={true} />
        )}
        <YBBreadCrumbComponent/>
        <Text ta="center" fw={700}>Welcome to ybFeed</Text>
        <Text ta="center">Choose a unique name for your feed :</Text>
        <form onSubmit={form.onSubmit((values) => handleFinish(values.feedName))}>
        <Container size="200">
            <TextInput mt="1em" size="xs" autoCapitalize='none' {...form.getInputProps('feedName')}/>
        </Container>
        </form>
        </>
    )
}