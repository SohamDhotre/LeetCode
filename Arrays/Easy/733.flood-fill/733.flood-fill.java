class Solution {
    public int[][] floodFill(int[][] image, int sr, int sc, int color) {
        if (image[sr][sc] == color) return image;
        int m=image.length, n=image[0].length;
        int [][]dir={
            {0, 1},
            {1, 0},
            {-1, 0},
            {0, -1}
        };
        Queue<int[]>q=new ArrayDeque<>();
        q.add(new int[]{sr, sc});
        int oldColor=image[sr][sc];
        image[sr][sc] = color;
        while(!q.isEmpty()){
            int []curPoint=q.poll();
            for(int []newDir:dir){
                int nsr=curPoint[0]+newDir[0];
                int nsc=curPoint[1]+newDir[1];
                if(inRange(nsr, nsc, m, n) && oldColor==image[nsr][nsc]){
                    image[nsr][nsc]=color;
                    q.offer(new int[]{nsr, nsc});
                }
            }
        }
        return image;
    }

    boolean inRange(int row, int col, int maxRow, int maxCol){
        if(row<0 || col<0 || row>=maxRow || col>=maxCol) return false;
        return true;
    }
}