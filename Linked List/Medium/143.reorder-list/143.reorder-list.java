/**
 * Definition for singly-linked list.
 * public class ListNode {
 *     int val;
 *     ListNode next;
 *     ListNode() {}
 *     ListNode(int val) { this.val = val; }
 *     ListNode(int val, ListNode next) { this.val = val; this.next = next; }
 * }
 */
class Solution {
    public void reorderList(ListNode head) {
        List<ListNode>list=getList(head);
        int left=1, right=list.size()-1;
        ListNode cur=head;
        while(left<=right){
            cur.next=list.get(right--);
            cur=cur.next;
            cur.next=null;
            // print(head, list.size(), "inside while");
            cur.next=list.get(left++);
            cur=cur.next;
            cur.next=null;
        }
        // print(head, list.size(), "outside while");
    }
    void print(ListNode head, int len, String status){
        ListNode cur=head;
        int count=len;
        System.out.print(status+" List: ");
        while(count>0){
            System.out.print(cur.val+"->");
            count--;
        }
        System.out.println(cur.val+"null");
    }
    List<ListNode> getList(ListNode head){
        List<ListNode>list=new ArrayList<>();
        ListNode cur=head;
        while(cur!=null){
            list.add(cur);
            cur=cur.next;
        }
        return list;
    }
}