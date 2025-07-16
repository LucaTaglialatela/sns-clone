import React, { useCallback, useEffect, useState } from "react";
import { PostItem } from "./PostItem";
import type { Post } from "../types/Post";
import { useAuth } from "../context/AuthContext";
import UserProfile from "./UserProfile";
import type { User } from "../types/User";

interface PostsListProps {
  posts: Post[];
  fetchAllPosts: () => void;
  selectedUser: User | null;
  setSelectedUser: (user: User | null) => void;
  setIsViewingProfile: (isViewingProfile: boolean) => void;
}

export const PostList: React.FC<PostsListProps> = ({
  posts,
  fetchAllPosts,
  selectedUser,
  setSelectedUser,
  setIsViewingProfile,
}) => {
  const { following, followUser, unfollowUser } = useAuth();
  const [editingPostId, setEditingPostId] = useState<string | null>(null);
  const [actionInProgress, setActionInProgress] = useState<string | null>(null);

  useEffect(() => {
    setIsViewingProfile(!!selectedUser);
  }, [selectedUser]);

  const onViewProfile = useCallback(async (userId: string) => {
    try {
      const response = await fetch(`/users/${userId}`);
      if (!response.ok)
        throw new Error(`HTTP error! status: ${response.status}`);
      const data: User = await response.json();
      setSelectedUser(data);
    } catch (e) {
      setSelectedUser(null);
    }
  }, []);

  const handleFollowToggle = async (
    targetUserId: string,
    isCurrentlyFollowing: boolean
  ) => {
    setActionInProgress(targetUserId);
    if (isCurrentlyFollowing) {
      await unfollowUser(targetUserId);
    } else {
      await followUser(targetUserId);
    }
    setActionInProgress(null);
  };

  const handleStartEdit = useCallback((post: Post) => {
    setEditingPostId(post.id);
  }, []);

  const handleCancelEdit = useCallback(() => {
    setEditingPostId(null);
  }, []);

  const handlePostUpdated = useCallback(() => {
    setEditingPostId(null);
    fetchAllPosts();
  }, [fetchAllPosts]);

  const handleDeletePost = useCallback(
    async (userId: string, postId: string) => {
      try {
        const response = await fetch(`users/${userId}/posts/${postId}`, {
          method: "DELETE",
        });
        if (!response.ok) {
          throw new Error("Failed to delete post.");
        }
        fetchAllPosts();
      } catch (err) {
        console.error(err);
        alert(
          err instanceof Error ? err.message : "An unknown error occurred."
        );
      }
    },
    [fetchAllPosts]
  );

  if (selectedUser) {
    const isFollowing = following.includes(selectedUser.id);
    const isButtonLoading = actionInProgress === selectedUser.id;

    return (
      <UserProfile
        user={selectedUser}
        isFollowing={isFollowing}
        isButtonLoading={isButtonLoading}
        setSelectedUser={setSelectedUser}
        handleFollowToggle={handleFollowToggle}
        returnButtonText={"Back to timeline"}
      />
    );
  }

  return (
    <div className="posts-container">
      {posts.length === 0 ? (
        <p>No posts yet. Be the first to create one!</p>
      ) : (
        <ul className="space-y-4">
          {posts.map((post) => (
            <PostItem
              key={post.id}
              post={post}
              editingPostId={editingPostId}
              onStartEdit={handleStartEdit}
              onCancelEdit={handleCancelEdit}
              onPostUpdated={handlePostUpdated}
              onDelete={handleDeletePost}
              onViewProfile={onViewProfile}
            />
          ))}
        </ul>
      )}
    </div>
  );
};

export default PostList;
