import React, { useState } from "react";
import PostForm from "./PostForm";
import type { Post } from "../types/Post";
import { calculateSHA256 } from "../utils/hash";

interface EditPostFormProps {
  post: Post;
  onPostUpdated: () => void;
  onCancel: () => void;
}

const EditPostForm: React.FC<EditPostFormProps> = ({
  post,
  onPostUpdated,
  onCancel,
}) => {
  const [text, setText] = useState<string>(post.text);
  const [file, setFile] = useState<File | null>(null);
  const [fileName, setFileName] = useState<string>(post.image);
  const [isSubmitting, setIsSubmitting] = useState<boolean>(false);
  const [error, setError] = useState<string | null>(null);

  const handleUpdate = async (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    setIsSubmitting(true);
    setError(null);

    try {
      let image: string = fileName;
      // If a file has been selected, upload it to S3
      if (file) {
        const hash = await calculateSHA256(file)
        // First we get the presigned url from our backend
        const presignResponse = await fetch("/presign", {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ fileName: file.name, fileType: file.type, fileHash: hash }),
        });
        if (!presignResponse.ok) throw new Error("Could not get upload URL.");

        const { url, key } = await presignResponse.json();

        // After receiving the presigned url, we use it to upload
        // the file to S3
        const s3Response = await fetch(url, {
          method: "PUT",
          headers: { "Content-Type": file.type },
          body: file,
        });
        if (!s3Response.ok) throw new Error("Failed to upload file to S3.");

        // We set the value of image to the presigned url key
        image = key;
      }

      const updatePostResponse = await fetch(
        `/users/${post.user_id}/posts/${post.id}`,
        {
          method: "PUT",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({
            text,
            image,
          }),
        }
      );

      if (!updatePostResponse.ok) {
        throw new Error(
          `Error: ${updatePostResponse.status} ${updatePostResponse.statusText}`
        );
      }

      onPostUpdated();
    } catch (err) {
      setError(
        err instanceof Error ? err.message : "An unknown error occurred."
      );
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <PostForm
      formId={`edit-post-${post.id}`}
      text={text}
      onTextChange={setText}
      file={file}
      onFileChange={setFile}
      fileName={fileName}
      onFileNameChange={setFileName}
      isSubmitting={isSubmitting}
      onSubmit={handleUpdate}
      error={error}
      submitButtonText="Save"
      onCancel={onCancel}
    />
  );
};

export default EditPostForm;
