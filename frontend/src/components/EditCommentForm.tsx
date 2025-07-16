import React, { useState } from "react";
import type { Comment } from "../types/Comment";
import CommentForm from "./CommentForm";

interface EditCommentFormProps {
  postId: string;
  comment: Comment;
  onCommentUpdated: () => void;
  onCancel: () => void;
}

const EditCommentForm: React.FC<EditCommentFormProps> = ({
  postId,
  comment,
  onCommentUpdated,
  onCancel,
}) => {
  const [text, setText] = useState<string>(comment.text);
  const [isSubmitting, setIsSubmitting] = useState<boolean>(false);
  const [error, setError] = useState<string | null>(null);

  const handleUpdate = async (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    setIsSubmitting(true);
    setError(null);

    try {
      const updatePostResponse = await fetch(
        `/users/${comment.user_id}/posts/${postId}/comments/${comment.id}`,
        {
          method: "PUT",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ text }),
        }
      );

      if (!updatePostResponse.ok) {
        throw new Error(
          `Error: ${updatePostResponse.status} ${updatePostResponse.statusText}`
        );
      }

      onCommentUpdated();
    } catch (err) {
      setError(
        err instanceof Error ? err.message : "An unknown error occurred."
      );
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <CommentForm
      text={text}
      setText={setText}
      onSubmit={handleUpdate}
      error={error}
      isSubmitting={isSubmitting}
      submitButtonText={"Save"}
      onCancel={onCancel}
    />
  );
};

export default EditCommentForm;
