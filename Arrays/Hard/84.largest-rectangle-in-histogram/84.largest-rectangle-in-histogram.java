class Solution {
    public int largestRectangleArea(int[] heights) {
        int n=heights.length, maxArea = 0;
        Deque<Integer>st=new ArrayDeque<>();
        for(int i=0;i<=n;i++){
            int curHeight=(i==n)?0:heights[i];
            
            while(!st.isEmpty() && heights[st.peek()]>curHeight){
                int h = heights[st.pop()];
                int left=st.isEmpty()?-1:st.peek();
                int width=i-left-1;
                maxArea=Math.max(maxArea, h*width);
            }
            st.push(i);
        }
        return maxArea;
    }
}