import { useCallback, useEffect, useState } from "react";
import type { Post } from "../types/Post";
import type { User } from "../types/User";
import PostList from "./PostList";
import { useAuth } from "../context/AuthContext";

interface UserProfileProps {
  user: User | null;
  isFollowing: boolean;
  isButtonLoading: boolean;
  setSelectedUser: (user: User | null) => void;
  handleFollowToggle: (
    targetUserId: string,
    isCurrentlyFollowing: boolean
  ) => Promise<void>;
  returnButtonText: string;
}

const UserProfile: React.FC<UserProfileProps> = ({
  user,
  isFollowing,
  isButtonLoading,
  setSelectedUser,
  handleFollowToggle,
  returnButtonText,
}) => {
  const { userId } = useAuth();
  const [userPosts, setUserPosts] = useState<Post[]>([]);
  const [isPostsLoading, setIsPostsLoading] = useState(false);

  const fetchUserPosts = useCallback(async () => {
    if (!user) return;
    setIsPostsLoading(true);
    try {
      const response = await fetch(`/posts/${user.id}`);
      if (!response.ok)
        throw new Error(`HTTP error! status: ${response.status}`);
      const data: Post[] = await response.json();
      setUserPosts(data);
    } catch (e) {
      console.error("Failed to fetch user posts:", e);
      setUserPosts([]);
    } finally {
      setIsPostsLoading(false);
    }
  }, []);

  useEffect(() => {
    if (!user) {
      setUserPosts([]);
      return;
    }

    fetchUserPosts();
  }, [user]);

  if (!user) {
    return <></>;
  }

  return (
    <div>
      <button
        onClick={() => setSelectedUser(null)}
        className="cursor-pointer text-sm font-medium mb-4 hover:text-gray-800 hover:underline"
      >
        &larr; {returnButtonText}
      </button>
      <div className="p-4 shadow-sm mb-6 flex justify-between items-center">
        <div className="flex items-center">
          <img
            src={user.picture}
            alt="user's picture"
            className="max-w-full h-auto mr-6 max-h-[300px] rounded-full inset-shadow-sm"
          />
          <div>
            <h2 className="text-2xl font-bold">{user.name}</h2>
            <p className="text-sm text-gray-600">{user.email}</p>
          </div>
        </div>
        {user.id !== userId && (
          <button
            onClick={() => handleFollowToggle(user.id, isFollowing)}
            disabled={isButtonLoading}
            className={`w-28 text-center px-4 py-1.5 cursor-pointer text-sm font-semibold rounded-full transition-colors duration-200 disabled:opacity-50 ${
              isFollowing
                ? "bg-white text-gray-700 border border-gray-300 hover:bg-gray-100"
                : "bg-black text-white hover:bg-gray-800"
            }`}
          >
            {isButtonLoading ? "..." : isFollowing ? "Following" : "Follow"}
          </button>
        )}
      </div>
      <h3 className="text-xl font-bold mb-4">{`${user.name}'s posts`}</h3>
      {isPostsLoading ? (
        <p className="text-gray-500">Loading posts...</p>
      ) : (
        <PostList
          posts={userPosts}
          fetchAllPosts={fetchUserPosts}
          setIsViewingProfile={() => {}}
          selectedUser={null}
          setSelectedUser={setSelectedUser}
        />
      )}
    </div>
  );
};

export default UserProfile;
