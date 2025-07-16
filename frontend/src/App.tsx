import { useCallback, useMemo, useState } from "react";
import CreatePost from "./components/CreatePostForm";
import PostList from "./components/PostList";
import UserList from "./components/UserList";
import { AuthProvider, useAuth } from "./context/AuthContext";
import SignIn from "./assets/sign_in.svg";
import SignOut from "./assets/sign_out.svg";
import type { User } from "./types/User";
import Logo from "./assets/logo.png";
import { SSEProvider, useSSE } from "./context/SSEContext";

type ActiveView = "timeline" | "users";
type TimelineView = "global" | "personal";

function App() {
  const { isAuthenticated, userId, userName, following, isLoading } = useAuth();
  const { posts } = useSSE();

  const [activeView, setActiveView] = useState<ActiveView>("timeline");
  const [activeTimeline, setActiveTimeline] = useState<TimelineView>("global");
  const [isViewingProfile, setIsViewingProfile] = useState<boolean>(false);
  const [selectedUser, setSelectedUser] = useState<User | null>(null);

  const closeProfile = useCallback(() => {
    setIsViewingProfile(false);
    setSelectedUser(null);
  }, []);

  const displayedPosts = useMemo(() => {
    if (activeTimeline === "personal") {
      const feedUserIds = new Set([userId, ...following]);
      return posts.filter((post) => feedUserIds.has(post.user_id));
    }
    return posts;
  }, [activeTimeline, posts, userId, following]);

  if (isLoading) {
    return (
      <div className="h-screen w-full flex items-center justify-center bg-gray-100">
        Loading...
      </div>
    );
  }

  if (!isAuthenticated) {
    return (
      <div className="h-screen w-full flex items-center justify-center bg-gray-100">
        <div className="text-center p-8 bg-white rounded-lg shadow-lg">
          <img src={Logo} alt="logo" className="h-10 w-10 block mx-auto mb-4" />
          <p className="mb-2 text-sm font-medium text-gray-600">
            Ready to share your thoughts with the world ?
          </p>
          <div className="flex justify-center cursor-pointer hover:opacity-75">
            <a
              href={`/auth/google/login`}
              className="flex text-sm font-medium"
            >
              <img src={SignIn} alt="sign in" className="h-5 w-5" />
              Sign in with Google
            </a>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="flex flex-col h-screen max-w-4xl mx-auto bg-white shadow-lg">
      <header className="p-6 border-b border-gray-200 flex-shrink-0">
        <div className="flex justify-between items-center">
          <img
            src={Logo}
            alt="logo"
            className="h-10 w-10 cursor-pointer hover:opacity-75 transition-opacity duration-100"
            onClick={() => {
              closeProfile();
              setActiveView("timeline");
            }}
          />
          <h1 className="text-m font-bold">{`Loggedddddddddd in as ${userName}`}</h1>
          <div className="cursor-pointer hover:opacity-75">
            <a
              href={`/auth/logout`}
              className="flex text-sm font-medium"
            >
              <img src={SignOut} alt="sign out" className="h-5 w-5" />
              Sign out
            </a>
          </div>
        </div>
      </header>

      <main className="flex-1 flex flex-col overflow-y-hidden">
        <div className="flex border-b border-gray-200">
          <button
            onClick={() => {
              closeProfile();
              setActiveView("timeline");
            }}
            className={`flex-1 py-4 text-center cursor-pointer font-medium transition-colors duration-200 focus:outline-none ${
              activeView === "timeline"
                ? "border-b-2"
                : "border-b-2 border-transparent text-gray-500 hover:bg-gray-100"
            }`}
          >
            Timeline
          </button>
          <button
            onClick={() => {
              closeProfile();
              setActiveView("users");
            }}
            className={`flex-1 py-4 text-center cursor-pointer font-medium transition-colors duration-200 focus:outline-none ${
              activeView === "users"
                ? "border-b-2"
                : "border-b-2 border-transparent text-gray-500 hover:bg-gray-100"
            }`}
          >
            Users
          </button>
        </div>

        <div className="flex-1 overflow-y-auto p-6 bg-gray-50">
          {activeView === "timeline" && (
            <div>
              {!isViewingProfile && (
                <CreatePost onPostCreated={() => {}} />
              )}
              <div className="mt-6">
                {!isViewingProfile && (
                  <div className="flex mb-4 items-center">
                    <h1 className="text-xl font-bold mr-4">Timeline</h1>
                    <button
                      onClick={() => setActiveTimeline("global")}
                      className={`px-4 py-2 cursor-pointer text-sm font-bold rounded-l-lg ${
                        activeTimeline === "global"
                          ? "bg-black text-white"
                          : "bg-gray-200"
                      }`}
                    >
                      Global
                    </button>
                    <button
                      onClick={() => setActiveTimeline("personal")}
                      className={`px-4 py-2 cursor-pointer text-sm font-bold rounded-r-lg ${
                        activeTimeline === "personal"
                          ? "bg-black text-white"
                          : "bg-gray-200"
                      }`}
                    >
                      Personal
                    </button>
                  </div>
                )}
                <PostList
                  posts={displayedPosts}
                  fetchAllPosts={() => {}}
                  selectedUser={selectedUser}
                  setSelectedUser={setSelectedUser}
                  setIsViewingProfile={setIsViewingProfile}
                />
              </div>
            </div>
          )}
          {activeView === "users" && (
            <UserList
              selectedUser={selectedUser}
              setSelectedUser={setSelectedUser}
              setIsViewingProfile={setIsViewingProfile}
            />
          )}
        </div>
      </main>
    </div>
  );
}

function Root() {
  return (
    <AuthProvider>
      <SSEProvider>
        <App />
      </SSEProvider>
    </AuthProvider>
  );
}

export default Root;
