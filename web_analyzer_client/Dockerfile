# Use the official nginx image
FROM nginx:alpine
# Remove default nginx index.html
RUN rm /usr/share/nginx/html/*

COPY index.html /usr/share/nginx/html/

# Expose port 80
EXPOSE 80
