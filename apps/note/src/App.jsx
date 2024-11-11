import "./App.css";
import {
  QueryClient,
  QueryClientProvider,
  useQuery,
} from "@tanstack/react-query";

// Create a client
const queryClient = new QueryClient();

function App() {
  return (
    // Provide the client to your App
    <QueryClientProvider client={queryClient}>
      <Notes />
    </QueryClientProvider>
  );
}

function Notes() {
  const { isPending, isError, data, error } = useQuery({
    queryKey: "notes",
    queryFn: async () => {
      const res = await fetch("/api/notes");
      const notes = await res.json();
      return notes;
    },
  });

  if (isPending) {
    return <span>Loading...</span>;
  }

  if (isError) {
    return <span>Error: {error.message}</span>;
  }

  return (
    <ul>
      {data.map((note) => (
        <li key={note.id}>{note.text}</li>
      ))}
    </ul>
  );
}

export default App;
